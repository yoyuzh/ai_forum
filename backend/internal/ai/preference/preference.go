// Package preference reads AI tag preference configuration.
package preference

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Preference struct {
	AgentID int64   `db:"ai_agent_id"`
	TagType string  `db:"tag_type"`
	TagName string  `db:"tag_name"`
	Weight  float64 `db:"weight"`
}

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) ListByAgent(ctx context.Context, agentID int64) ([]Preference, error) {
	var prefs []Preference
	if err := r.db.SelectContext(ctx, &prefs, `
		SELECT ai_agent_id, tag_type, tag_name, weight
		FROM ai_agent_tag_preferences
		WHERE ai_agent_id = ?
		ORDER BY id`, agentID); err != nil {
		return nil, fmt.Errorf("list ai preferences: %w", err)
	}
	return prefs, nil
}
