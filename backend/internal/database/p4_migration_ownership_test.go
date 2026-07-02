package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestP4MigrationsDoNotRecreateOwnedTables(t *testing.T) {
	dir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	var p4Files []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		if name >= "000005_" && name <= "000012_zzzzzz.up.sql" {
			p4Files = append(p4Files, name)
			body, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			sql := strings.ToLower(string(body))
			for _, forbidden := range []string{
				"create table users",
				"create table outbox_events",
				"create table processed_events",
			} {
				if strings.Contains(sql, forbidden) {
					t.Fatalf("%s must not contain %q", name, forbidden)
				}
			}
		}
	}

	if len(p4Files) != 8 {
		t.Fatalf("expected 8 P4 up migrations 000005..000012, got %d: %v", len(p4Files), p4Files)
	}
}
