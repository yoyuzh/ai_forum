package rbac

import (
	"database/sql"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/jmoiron/sqlx"
)

type sqlxAdapter struct {
	db *sqlx.DB
}

func NewSQLXAuthorizer(modelPath string, db *sqlx.DB) (*Authorizer, error) {
	e, err := casbin.NewEnforcer(modelPath, &sqlxAdapter{db: db})
	if err != nil {
		return nil, err
	}
	return &Authorizer{enforcer: e}, nil
}

func (a *sqlxAdapter) LoadPolicy(m model.Model) error {
	var rows []casbinRule
	if err := a.db.Select(&rows, `SELECT ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule ORDER BY id`); err != nil {
		return err
	}
	for _, row := range rows {
		rule := compactRule(row.V0, row.V1, row.V2, row.V3, row.V4, row.V5)
		if err := m.AddPolicy(policySection(row.PType), row.PType, rule); err != nil {
			return err
		}
	}
	return nil
}

func (a *sqlxAdapter) SavePolicy(m model.Model) error {
	tx, err := a.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM casbin_rule`); err != nil {
		return err
	}
	for sec, assertions := range m {
		if sec != "p" && sec != "g" {
			continue
		}
		for ptype, assertion := range assertions {
			for _, rule := range assertion.Policy {
				if err := insertRule(tx, ptype, rule); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit()
}

func (a *sqlxAdapter) AddPolicy(_ string, ptype string, rule []string) error {
	return insertRule(a.db, ptype, rule)
}

func (a *sqlxAdapter) RemovePolicy(_ string, ptype string, rule []string) error {
	_, err := a.db.Exec(`DELETE FROM casbin_rule WHERE ptype=? AND v0<=>? AND v1<=>? AND v2<=>? AND v3<=>? AND v4<=>? AND v5<=>?`, ruleArg(ptype, rule)...)
	return err
}

func (a *sqlxAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return fmt.Errorf("filtered policy removal not implemented for %s/%s/%d/%v", sec, ptype, fieldIndex, fieldValues)
}

type casbinRule struct {
	PType string  `db:"ptype"`
	V0    *string `db:"v0"`
	V1    *string `db:"v1"`
	V2    *string `db:"v2"`
	V3    *string `db:"v3"`
	V4    *string `db:"v4"`
	V5    *string `db:"v5"`
}

func insertRule(db interface {
	Exec(query string, args ...any) (sql.Result, error)
}, ptype string, rule []string) error {
	_, err := db.Exec(`INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES (?, ?, ?, ?, ?, ?, ?)`, ruleArg(ptype, rule)...)
	return err
}

func ruleArg(ptype string, rule []string) []any {
	args := []any{ptype}
	for i := 0; i < 6; i++ {
		if i < len(rule) {
			args = append(args, rule[i])
		} else {
			args = append(args, nil)
		}
	}
	return args
}

func compactRule(values ...*string) []string {
	rule := make([]string, 0, len(values))
	for _, value := range values {
		if value == nil {
			break
		}
		rule = append(rule, *value)
	}
	return rule
}

func policySection(ptype string) string {
	if len(ptype) > 0 && ptype[0] == 'g' {
		return "g"
	}
	return "p"
}
