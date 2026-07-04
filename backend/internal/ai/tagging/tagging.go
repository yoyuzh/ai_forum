// Package tagging generates post tags for AI decisions.
package tagging

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/ai/modelclient"
	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/event"
	forumpost "ai-forum/backend/internal/forum/post"
	forumtag "ai-forum/backend/internal/forum/tag"
	"ai-forum/backend/internal/outbox"
)

type Post struct {
	ID      int64
	Title   string
	Content string
}

type Tag struct {
	Type string
	Name string
}

type Tagger interface {
	Tag(context.Context, Post) []Tag
}

type RuleTagger struct{}

type ModelTagger struct {
	client   modelclient.Client
	fallback Tagger
}

type Tags struct {
	Topic   []string `json:"topic"`
	Intent  []string `json:"intent"`
	Emotion []string `json:"emotion"`
	Debate  []string `json:"debate"`
	Risk    []string `json:"risk"`
}

type PostReader interface {
	GetPost(ctx context.Context, postID int64) (Post, error)
}

type TagWriter interface {
	ReplaceTags(ctx context.Context, postID int64, tags []Tag) error
}

type OutboxAppender interface {
	AppendPostTagged(ctx context.Context, postID int64, tags []Tag) error
}

type Handler struct {
	posts  PostReader
	tags   TagWriter
	outbox OutboxAppender
	tagger Tagger
}

type SQLHandler struct {
	db     *sqlx.DB
	tagger Tagger
}

func NewHandler(posts PostReader, tags TagWriter, outbox OutboxAppender, tagger Tagger) *Handler {
	return &Handler{posts: posts, tags: tags, outbox: outbox, tagger: tagger}
}

func NewSQLHandler(db *sqlx.DB, tagger Tagger) *SQLHandler {
	return &SQLHandler{db: db, tagger: tagger}
}

func NewModelTagger(client modelclient.Client, fallback Tagger) *ModelTagger {
	if fallback == nil {
		fallback = RuleTagger{}
	}
	return &ModelTagger{client: client, fallback: fallback}
}

func (h *Handler) HandleTagPost(ctx context.Context, postID int64) error {
	post, err := h.posts.GetPost(ctx, postID)
	if err != nil {
		return err
	}
	tags := h.tagger.Tag(ctx, post)
	if err := h.tags.ReplaceTags(ctx, postID, tags); err != nil {
		return err
	}
	return h.outbox.AppendPostTagged(ctx, postID, tags)
}

func (h *SQLHandler) HandleTagPost(ctx context.Context, postID int64) error {
	postRepo := forumpost.NewSQLRepository()
	tagRepo := forumtag.NewSQLRepository()
	return database.RunInTx(ctx, h.db, func(tx *sqlx.Tx) error {
		p, err := postRepo.Get(ctx, tx, postID)
		if err != nil {
			return err
		}
		tags := h.tagger.Tag(ctx, Post{ID: p.ID, Title: p.Title, Content: p.Content})
		forumTags := make([]forumtag.Tag, 0, len(tags))
		for _, tag := range tags {
			forumTags = append(forumTags, forumtag.Tag{PostID: postID, Type: tag.Type, Name: tag.Name})
		}
		if err := tagRepo.Replace(ctx, tx, postID, forumTags); err != nil {
			return err
		}
		return outbox.Append(ctx, tx, outbox.Event{
			EventType:     event.PostTagged,
			AggregateType: "post",
			AggregateID:   postID,
			Payload:       map[string]any{"post_id": postID, "tags": tags},
		})
	})
}

func (RuleTagger) Tag(_ context.Context, post Post) []Tag {
	text := strings.ToLower(post.Title + " " + post.Content)
	return []Tag{
		{Type: "topic", Name: topicTag(text)},
		{Type: "intent", Name: intentTag(text)},
		{Type: "emotion", Name: emotionTag(text)},
		{Type: "debate", Name: debateTag(text)},
		{Type: "risk", Name: riskTag(text)},
	}
}

func (t *ModelTagger) Tag(ctx context.Context, post Post) []Tag {
	if t.client == nil {
		return t.fallback.Tag(ctx, post)
	}
	temp := 0.1
	raw, err := t.client.Generate(ctx, modelclient.Request{
		SystemPrompt: tagSystemPrompt,
		Prompt:       buildTagPrompt(post),
		MaxTokens:    300,
		Temperature:  &temp,
		TaskType:     "tag_post",
		PostID:       post.ID,
	})
	if err != nil {
		return t.fallback.Tag(ctx, post)
	}
	tags := flattenTags(ParseTags(raw))
	if len(tags) == 0 {
		return t.fallback.Tag(ctx, post)
	}
	return tags
}

func ParseTags(raw string) Tags {
	var tags Tags
	if err := json.Unmarshal([]byte(stripMarkdownFence(raw)), &tags); err != nil {
		return Tags{Intent: []string{"求建议"}, Debate: []string{"争议性低"}, Risk: []string{"正常"}}
	}
	return filterValidTags(tags)
}

func flattenTags(tags Tags) []Tag {
	out := make([]Tag, 0, len(tags.Topic)+len(tags.Intent)+len(tags.Emotion)+len(tags.Debate)+len(tags.Risk))
	for _, name := range tags.Topic {
		out = append(out, Tag{Type: "topic", Name: name})
	}
	for _, name := range tags.Intent {
		out = append(out, Tag{Type: "intent", Name: name})
	}
	for _, name := range tags.Emotion {
		out = append(out, Tag{Type: "emotion", Name: name})
	}
	for _, name := range tags.Debate {
		out = append(out, Tag{Type: "debate", Name: name})
	}
	for _, name := range tags.Risk {
		out = append(out, Tag{Type: "risk", Name: name})
	}
	return out
}

func stripMarkdownFence(raw string) string {
	s := strings.TrimSpace(raw)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func filterValidTags(tags Tags) Tags {
	return Tags{
		Topic:   filterAllowed(tags.Topic, topicCandidates),
		Intent:  filterAllowed(tags.Intent, intentCandidates),
		Emotion: filterAllowed(tags.Emotion, emotionCandidates),
		Debate:  filterAllowed(tags.Debate, debateCandidates),
		Risk:    filterAllowed(tags.Risk, riskCandidates),
	}
}

func filterAllowed(values []string, allowed map[string]bool) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		if !allowed[value] || seen[value] || len(out) >= 3 {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func buildTagPrompt(post Post) string {
	return "帖子标题：" + post.Title + "\n帖子正文：" + post.Content + "\n\n" + tagUserPrompt
}

const tagSystemPrompt = "你是一个论坛帖子分析助手。分析用户帖子，从固定候选集中选出匹配的标签，输出JSON，不要输出其他内容。"

const tagUserPrompt = `请从以下候选集中选出匹配的标签，每类选0~3个，输出JSON：

topic候选：学习规划、软件工程、大学生活、校园、找工作、实习、职场、职业选择、创业、收入、投资、管理、效率、技术、编程、项目、社会现象、理想、意义、价值观、重大决策、人际关系、情感

intent候选：求建议、求评价、求共鸣、情绪倾诉、技术分析、讨论观点、可行性分析、选择困难、经验分享

emotion候选：焦虑、迷茫、压力、自我怀疑、热血、孤独、自嘲、无奈、愤怒、开心

debate候选：争议性高、争议性中、争议性低、价值权衡、保守vs创新、技术选型

risk候选：正常、敏感、高风险

输出格式：
{
  "topic": [],
  "intent": [],
  "emotion": [],
  "debate": [],
  "risk": []
}`

var topicCandidates = candidateSet("学习规划", "软件工程", "大学生活", "校园", "找工作", "实习", "职场", "职业选择", "创业", "收入", "投资", "管理", "效率", "技术", "编程", "项目", "社会现象", "理想", "意义", "价值观", "重大决策", "人际关系", "情感")
var intentCandidates = candidateSet("求建议", "求评价", "求共鸣", "情绪倾诉", "技术分析", "讨论观点", "可行性分析", "选择困难", "经验分享")
var emotionCandidates = candidateSet("焦虑", "迷茫", "压力", "自我怀疑", "热血", "孤独", "自嘲", "无奈", "愤怒", "开心")
var debateCandidates = candidateSet("争议性高", "争议性中", "争议性低", "价值权衡", "保守vs创新", "技术选型")
var riskCandidates = candidateSet("正常", "敏感", "高风险")

func candidateSet(values ...string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func topicTag(text string) string {
	if strings.Contains(text, "ai") {
		return "ai"
	}
	if strings.Contains(text, "debate") {
		return "debate"
	}
	return "general"
}

func intentTag(text string) string {
	if strings.Contains(text, "?") || strings.Contains(text, "should") {
		return "question"
	}
	return "discussion"
}

func emotionTag(text string) string {
	if strings.Contains(text, "worried") || strings.Contains(text, "concern") {
		return "concerned"
	}
	return "neutral"
}

func debateTag(text string) string {
	if strings.Contains(text, "debate") || strings.Contains(text, "should") {
		return "high"
	}
	return "low"
}

func riskTag(text string) string {
	if strings.Contains(text, "risk") || strings.Contains(text, "safety") {
		return "sensitive"
	}
	return "normal"
}
