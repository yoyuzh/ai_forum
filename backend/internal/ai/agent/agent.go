// Package agent reads AI agent profiles and tag preferences.
package agent

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Agent is an enabled AI agent profile used by the decision worker.
type Agent struct {
	ID             int64
	Name           string
	SystemPrompt   string
	ReplyThreshold float64
	ActivityLevel  float64
	AllowAutoReply bool
	AllowMention   bool
	AllowFollowup  bool
	Fallback       bool
	Preferences    []Preference
}

// Preference is an agent tag preference weight.
type Preference struct {
	TagType string
	TagName string
	Weight  float64
}

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) ListEnabledWithPreferences(ctx context.Context) ([]Agent, error) {
	var rows []struct {
		ID             int64   `db:"id"`
		Name           string  `db:"name"`
		SystemPrompt   string  `db:"system_prompt"`
		ReplyThreshold float64 `db:"reply_threshold"`
		ActivityLevel  float64 `db:"activity_level"`
		AllowAutoReply bool    `db:"allow_auto_reply"`
		AllowMention   bool    `db:"allow_mention"`
		AllowFollowup  bool    `db:"allow_followup"`
		Fallback       bool    `db:"is_fallback"`
	}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, name, COALESCE(system_prompt, '') AS system_prompt, reply_threshold, activity_level, allow_auto_reply, allow_mention, allow_followup, is_fallback
		FROM ai_agents
		WHERE enabled = TRUE
		ORDER BY id`); err != nil {
		return nil, fmt.Errorf("list enabled ai agents: %w", err)
	}
	agents := make([]Agent, 0, len(rows))
	byID := make(map[int64]int, len(rows))
	for _, row := range rows {
		byID[row.ID] = len(agents)
		agents = append(agents, Agent{
			ID:             row.ID,
			Name:           row.Name,
			SystemPrompt:   row.SystemPrompt,
			ReplyThreshold: row.ReplyThreshold,
			ActivityLevel:  row.ActivityLevel,
			AllowAutoReply: row.AllowAutoReply,
			AllowMention:   row.AllowMention,
			AllowFollowup:  row.AllowFollowup,
			Fallback:       row.Fallback,
		})
	}
	if len(agents) == 0 {
		return agents, nil
	}
	var prefs []struct {
		AgentID int64   `db:"ai_agent_id"`
		TagType string  `db:"tag_type"`
		TagName string  `db:"tag_name"`
		Weight  float64 `db:"weight"`
	}
	if err := r.db.SelectContext(ctx, &prefs, `
		SELECT ai_agent_id, tag_type, tag_name, weight
		FROM ai_agent_tag_preferences
		ORDER BY ai_agent_id, id`); err != nil {
		return nil, fmt.Errorf("list ai agent preferences: %w", err)
	}
	for _, pref := range prefs {
		i, ok := byID[pref.AgentID]
		if !ok {
			continue
		}
		agents[i].Preferences = append(agents[i].Preferences, Preference{
			TagType: pref.TagType,
			TagName: pref.TagName,
			Weight:  pref.Weight,
		})
	}
	return agents, nil
}
