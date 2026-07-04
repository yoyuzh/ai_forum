// Package decision computes AI reply willingness and selection.
package decision

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	aiagent "ai-forum/backend/internal/ai/agent"
	"ai-forum/backend/internal/ai/modelclient"
	forumtag "ai-forum/backend/internal/forum/tag"
)

const (
	DecisionReply         = "REPLY"
	DecisionIgnore        = "IGNORE"
	DecisionFallback      = "FALLBACK"
	DefaultAutoReplyLimit = 3
)

// TagScore is an agent preference weight for a post tag.
type TagScore struct {
	Type   string
	Weight float64
}

// ScoreInput contains formula inputs for §11.2.
type ScoreInput struct {
	Tags             []TagScore
	ActivityScore    float64
	RiskPenalty      float64
	FrequencyPenalty float64
}

// AgentScore is a computed score plus selection metadata.
type AgentScore struct {
	AgentID   int64
	Score     float64
	Threshold float64
	Fallback  bool
	Decision  string
}

type Agent struct {
	ID             int64
	Name           string
	Persona        string
	ReplyThreshold float64
	ActivityLevel  float64
	AllowAutoReply bool
	Fallback       bool
	Preferences    []Preference
}

type Preference struct {
	TagType string
	TagName string
	Weight  float64
}

type PostTag struct {
	Type string
	Name string
}

type Log struct {
	PostID           int64
	AgentID          int64
	TriggerType      string
	WillingnessScore float64
	ThresholdValue   float64
	Decision         string
	Reason           string
	HitTags          []PostTag
}

type AgentReader interface {
	ListEnabledAgents(context.Context) ([]Agent, error)
}

type TagReader interface {
	ListPostTags(context.Context, int64) ([]PostTag, error)
}

type DecisionLogger interface {
	WriteDecisionLog(context.Context, Log) error
}

type ReplyEnqueuer interface {
	EnqueueAutoGenerateAIReply(ctx context.Context, postID, agentID int64) error
}

type WillingnessScorer interface {
	ScoreWillingness(context.Context, WillingnessInput) (float64, error)
}

type WillingnessInput struct {
	PostID        int64
	Agent         Agent
	Tags          []PostTag
	FallbackScore float64
}

type Handler struct {
	agents   AgentReader
	tags     TagReader
	logs     DecisionLogger
	enqueuer ReplyEnqueuer
	scorer   WillingnessScorer
}

type SQLHandler struct {
	db       *sqlx.DB
	agents   *aiagent.SQLRepository
	tags     *forumtag.SQLRepository
	enqueuer ReplyEnqueuer
	scorer   WillingnessScorer
}

type ModelWillingnessScorer struct {
	client modelclient.Client
}

func NewHandler(agents AgentReader, tags TagReader, logs DecisionLogger, enqueuer ReplyEnqueuer) *Handler {
	return &Handler{agents: agents, tags: tags, logs: logs, enqueuer: enqueuer}
}

func NewSQLHandler(db *sqlx.DB, enqueuer ReplyEnqueuer) *SQLHandler {
	return &SQLHandler{
		db:       db,
		agents:   aiagent.NewSQLRepository(db),
		tags:     forumtag.NewSQLRepository(),
		enqueuer: enqueuer,
	}
}

func (h *Handler) SetWillingnessScorer(scorer WillingnessScorer) {
	h.scorer = scorer
}

func (h *SQLHandler) SetWillingnessScorer(scorer WillingnessScorer) {
	h.scorer = scorer
}

func NewModelWillingnessScorer(client modelclient.Client) *ModelWillingnessScorer {
	return &ModelWillingnessScorer{client: client}
}

func (h *Handler) HandleDecideAIReply(ctx context.Context, postID int64) error {
	agents, err := h.agents.ListEnabledAgents(ctx)
	if err != nil {
		return err
	}
	postTags, err := h.tags.ListPostTags(ctx, postID)
	if err != nil {
		return err
	}
	if HasHighRisk(postTags) {
		for _, agent := range agents {
			if !agent.AllowAutoReply {
				continue
			}
			if err := h.logs.WriteDecisionLog(ctx, Log{
				PostID:      postID,
				AgentID:     agent.ID,
				TriggerType: "AUTO",
				Decision:    DecisionIgnore,
				Reason:      "high risk tag",
				HitTags:     []PostTag{{Type: "risk", Name: "高风险"}},
			}); err != nil {
				return err
			}
		}
		return nil
	}
	scores := make([]AgentScore, 0, len(agents))
	logs := make([]Log, 0, len(agents))
	for _, agent := range agents {
		if !agent.AllowAutoReply {
			continue
		}
		hitTags, tagScores := matchTags(agent.Preferences, postTags)
		score := WillingnessScore(ScoreInput{Tags: tagScores, ActivityScore: agent.ActivityLevel})
		score = scoreWithFallback(ctx, h.scorer, WillingnessInput{PostID: postID, Agent: agent, Tags: postTags, FallbackScore: score})
		scores = append(scores, AgentScore{AgentID: agent.ID, Score: score, Threshold: agent.ReplyThreshold, Fallback: agent.Fallback})
		logs = append(logs, Log{
			PostID:           postID,
			AgentID:          agent.ID,
			TriggerType:      "AUTO",
			WillingnessScore: score,
			ThresholdValue:   agent.ReplyThreshold,
			Decision:         DecisionIgnore,
			Reason:           "below threshold",
			HitTags:          hitTags,
		})
	}
	selected := SelectReplies(scores)
	selectedByID := map[int64]string{}
	selectedIDs := make([]int64, 0, len(selected))
	for _, agent := range selected {
		selectedByID[agent.AgentID] = agent.Decision
		selectedIDs = append(selectedIDs, agent.AgentID)
	}
	for _, log := range logs {
		if decision, ok := selectedByID[log.AgentID]; ok {
			log.Decision = decision
			log.Reason = "selected"
		}
		if err := h.logs.WriteDecisionLog(ctx, log); err != nil {
			return err
		}
	}
	for _, agentID := range selectedIDs {
		if err := h.enqueuer.EnqueueAutoGenerateAIReply(ctx, postID, agentID); err != nil {
			return err
		}
	}
	return nil
}

func (h *SQLHandler) HandleDecideAIReply(ctx context.Context, postID int64) error {
	var existing int
	if err := h.db.GetContext(ctx, &existing, `SELECT COUNT(*) FROM decision_logs WHERE post_id = ? AND trigger_type = 'AUTO'`, postID); err != nil {
		return fmt.Errorf("count decision logs: %w", err)
	}
	if existing > 0 {
		return h.enqueueSelectedFromLogs(ctx, postID)
	}
	agents, err := h.agents.ListEnabledWithPreferences(ctx)
	if err != nil {
		return err
	}
	postTags, err := h.tags.List(ctx, h.db, postID)
	if err != nil {
		return err
	}
	logs, selected := buildDecisionLogsWithScorer(ctx, postID, convertAgents(agents), convertTags(postTags), h.scorer)
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin decision transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	for _, log := range logs {
		hitTags, err := MarshalHitTags(log.HitTags)
		if err != nil {
			return fmt.Errorf("marshal decision hit tags: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO decision_logs (
				post_id, comment_id, ai_agent_id, trigger_type, willingness_score,
				threshold_value, decision, reason, hit_tags
			) VALUES (?, NULL, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				willingness_score = VALUES(willingness_score),
				threshold_value = VALUES(threshold_value),
				decision = VALUES(decision),
				reason = VALUES(reason),
				hit_tags = VALUES(hit_tags)`,
			log.PostID, log.AgentID, log.TriggerType, log.WillingnessScore,
			log.ThresholdValue, log.Decision, log.Reason, hitTags); err != nil {
			return fmt.Errorf("write decision log: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit decision logs: %w", err)
	}
	committed = true
	for _, agentID := range selected {
		if err := h.enqueuer.EnqueueAutoGenerateAIReply(ctx, postID, agentID); err != nil {
			return err
		}
	}
	return nil
}

func (h *SQLHandler) enqueueSelectedFromLogs(ctx context.Context, postID int64) error {
	var agentIDs []int64
	if err := h.db.SelectContext(ctx, &agentIDs, `
		SELECT ai_agent_id
		FROM decision_logs
		WHERE post_id = ? AND trigger_type = 'AUTO' AND decision IN (?, ?)
		ORDER BY id`, postID, DecisionReply, DecisionFallback); err != nil {
		return fmt.Errorf("list selected decision logs: %w", err)
	}
	for _, agentID := range agentIDs {
		if err := h.enqueuer.EnqueueAutoGenerateAIReply(ctx, postID, agentID); err != nil {
			return err
		}
	}
	return nil
}

// WillingnessScore computes the §11.2 weighted score.
func WillingnessScore(in ScoreInput) float64 {
	return tagTypeScore(in.Tags, "topic")*0.35 +
		tagTypeScore(in.Tags, "intent")*0.25 +
		tagTypeScore(in.Tags, "emotion")*0.15 +
		tagTypeScore(in.Tags, "debate")*0.15 +
		in.ActivityScore*0.10 -
		in.RiskPenalty -
		in.FrequencyPenalty
}

// SelectReplies picks the highest willingness scores for auto replies.
func SelectReplies(agents []AgentScore) []AgentScore {
	if len(agents) == 0 {
		return nil
	}
	sorted := append([]AgentScore(nil), agents...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})
	limit := DefaultAutoReplyLimit
	if len(sorted) < limit {
		limit = len(sorted)
	}
	selected := make([]AgentScore, 0, limit)
	for _, agent := range sorted {
		if len(selected) >= limit {
			break
		}
		if agent.Score <= 0 {
			continue
		}
		agent.Decision = DecisionReply
		selected = append(selected, agent)
	}
	return selected
}

func scoreWithFallback(ctx context.Context, scorer WillingnessScorer, in WillingnessInput) float64 {
	if scorer == nil {
		return in.FallbackScore
	}
	score, err := scorer.ScoreWillingness(ctx, in)
	if err != nil || score < 0 || score > 1 {
		return in.FallbackScore
	}
	return score
}

func (s *ModelWillingnessScorer) ScoreWillingness(ctx context.Context, in WillingnessInput) (float64, error) {
	if s == nil || s.client == nil {
		return 0, fmt.Errorf("model willingness scorer unavailable")
	}
	temp := 0.1
	raw, err := s.client.Generate(ctx, modelclient.Request{
		SystemPrompt: "你是论坛AI角色的回复意愿评分器。只输出JSON，不要解释。",
		Prompt:       buildWillingnessPrompt(in),
		MaxTokens:    80,
		Temperature:  &temp,
		TaskType:     "decide_ai_reply",
		PostID:       in.PostID,
		AIAgentID:    in.Agent.ID,
		TriggerType:  "AUTO",
	})
	if err != nil {
		return 0, err
	}
	return parseWillingnessScore(raw)
}

func buildWillingnessPrompt(in WillingnessInput) string {
	var b strings.Builder
	b.WriteString("请根据这个AI角色的一句话性格和帖子标签，输出它主动回复该帖子的意愿分。\n")
	b.WriteString("满分是1.00，最低是0.00。不要使用百分制或10分制。\n")
	fmt.Fprintf(&b, "该AI的回复阈值是%.2f；如果你认为它应该主动回复，score应高于这个阈值。\n", in.Agent.ReplyThreshold)
	fmt.Fprintf(&b, "AI角色：id=%d name=%s persona=%s\n", in.Agent.ID, in.Agent.Name, oneLinePersona(in.Agent))
	b.WriteString("帖子标签：")
	for i, tag := range in.Tags {
		if i > 0 {
			b.WriteString("、")
		}
		fmt.Fprintf(&b, "%s=%s", tag.Type, tag.Name)
	}
	fmt.Fprintf(&b, "\n本地公式参考分：%.4f\n", in.FallbackScore)
	b.WriteString("只返回JSON：{\"score\":0.72}")
	return b.String()
}

func oneLinePersona(agent Agent) string {
	for _, line := range strings.Split(agent.Persona, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	if agent.Name != "" {
		return agent.Name
	}
	return strconv.FormatInt(agent.ID, 10)
}

func parseWillingnessScore(raw string) (float64, error) {
	var wrapped struct {
		Score *float64 `json:"score"`
	}
	if err := json.Unmarshal([]byte(stripJSONFence(raw)), &wrapped); err == nil && wrapped.Score != nil {
		return *wrapped.Score, nil
	}
	var score float64
	if err := json.Unmarshal([]byte(stripJSONFence(raw)), &score); err != nil {
		return 0, err
	}
	return score, nil
}

func stripJSONFence(raw string) string {
	s := strings.TrimSpace(raw)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func tagTypeScore(tags []TagScore, tagType string) float64 {
	var max, sum float64
	var count int
	for _, tag := range tags {
		if tag.Type != tagType {
			continue
		}
		if count == 0 || tag.Weight > max {
			max = tag.Weight
		}
		sum += tag.Weight
		count++
	}
	if count == 0 {
		return 0
	}
	return max*0.7 + (sum/float64(count))*0.3
}

func matchTags(preferences []Preference, tags []PostTag) ([]PostTag, []TagScore) {
	var hits []PostTag
	var scores []TagScore
	for _, tag := range tags {
		for _, pref := range preferences {
			if pref.TagType == tag.Type && pref.TagName == tag.Name {
				hits = append(hits, tag)
				scores = append(scores, TagScore{Type: tag.Type, Weight: pref.Weight})
			}
		}
	}
	return hits, scores
}

func buildDecisionLogs(postID int64, agents []Agent, postTags []PostTag) ([]Log, []int64) {
	return buildDecisionLogsWithScorer(context.Background(), postID, agents, postTags, nil)
}

func buildDecisionLogsWithScorer(ctx context.Context, postID int64, agents []Agent, postTags []PostTag, scorer WillingnessScorer) ([]Log, []int64) {
	if HasHighRisk(postTags) {
		logs := make([]Log, 0, len(agents))
		for _, agent := range agents {
			if !agent.AllowAutoReply {
				continue
			}
			logs = append(logs, Log{
				PostID:      postID,
				AgentID:     agent.ID,
				TriggerType: "AUTO",
				Decision:    DecisionIgnore,
				Reason:      "high risk tag",
				HitTags:     []PostTag{{Type: "risk", Name: "高风险"}},
			})
		}
		return logs, nil
	}
	scores := make([]AgentScore, 0, len(agents))
	logs := make([]Log, 0, len(agents))
	for _, agent := range agents {
		if !agent.AllowAutoReply {
			continue
		}
		hitTags, tagScores := matchTags(agent.Preferences, postTags)
		score := WillingnessScore(ScoreInput{Tags: tagScores, ActivityScore: agent.ActivityLevel})
		score = scoreWithFallback(ctx, scorer, WillingnessInput{PostID: postID, Agent: agent, Tags: postTags, FallbackScore: score})
		scores = append(scores, AgentScore{AgentID: agent.ID, Score: score, Threshold: agent.ReplyThreshold, Fallback: agent.Fallback})
		logs = append(logs, Log{
			PostID:           postID,
			AgentID:          agent.ID,
			TriggerType:      "AUTO",
			WillingnessScore: score,
			ThresholdValue:   agent.ReplyThreshold,
			Decision:         DecisionIgnore,
			Reason:           "below threshold",
			HitTags:          hitTags,
		})
	}
	selected := SelectReplies(scores)
	selectedByID := map[int64]string{}
	selectedIDs := make([]int64, 0, len(selected))
	for _, agent := range selected {
		selectedByID[agent.AgentID] = agent.Decision
		selectedIDs = append(selectedIDs, agent.AgentID)
	}
	for i := range logs {
		if decision, ok := selectedByID[logs[i].AgentID]; ok {
			logs[i].Decision = decision
			logs[i].Reason = "selected"
		}
	}
	return logs, selectedIDs
}

func convertAgents(rows []aiagent.Agent) []Agent {
	agents := make([]Agent, 0, len(rows))
	for _, row := range rows {
		prefs := make([]Preference, 0, len(row.Preferences))
		for _, pref := range row.Preferences {
			prefs = append(prefs, Preference{TagType: pref.TagType, TagName: pref.TagName, Weight: pref.Weight})
		}
		agents = append(agents, Agent{
			ID:             row.ID,
			Name:           row.Name,
			Persona:        row.SystemPrompt,
			ReplyThreshold: row.ReplyThreshold,
			ActivityLevel:  row.ActivityLevel,
			AllowAutoReply: row.AllowAutoReply,
			Fallback:       row.Fallback,
			Preferences:    prefs,
		})
	}
	return agents
}

func convertTags(rows []forumtag.Tag) []PostTag {
	tags := make([]PostTag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, PostTag{Type: row.Type, Name: row.Name})
	}
	return tags
}

func HasHighRisk(tags []PostTag) bool {
	for _, tag := range tags {
		if tag.Type == "risk" && tag.Name == "高风险" {
			return true
		}
	}
	return false
}

func MarshalHitTags(tags []PostTag) ([]byte, error) {
	if tags == nil {
		tags = []PostTag{}
	}
	return json.Marshal(tags)
}
