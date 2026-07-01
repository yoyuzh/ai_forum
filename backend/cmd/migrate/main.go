// Package main is a thin wrapper around the golang-migrate library that
// imports the mysql and file drivers so they register when run via `go run`.
//
// Why this exists: running the upstream migrate CLI via
// `go run github.com/.../cmd/migrate@v4.18.1` does NOT compile the driver
// subpackages, so `mysql://` is "unknown driver (forgotten import?)". The
// golang-migrate project documents wrapping the library with blank-imported
// drivers as the fix. This wrapper mirrors the upstream CLI's flag set so
// the Makefile's `-path`/`-database` usage is unchanged.
//
// Usage:
//
//	go run ./cmd/migrate -path backend/migrations -database "mysql://..." up
//	go run ./cmd/migrate -path backend/migrations -database "mysql://..." down
//	go run ./cmd/migrate -path backend/migrations -database "mysql://..." steps -2
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql" // register mysql://
	_ "github.com/golang-migrate/migrate/v4/source/file"    // register file://
)

func main() {
	path := flag.String("path", "", "path to migrations directory (file:// scheme is added automatically)")
	database := flag.String("database", "", "database DSN, e.g. mysql://user:pass@tcp(host:port)/db")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: migrate -path DIR -database DSN <up|down|steps N|force V|version>")
		os.Exit(2)
	}
	if *path == "" || *database == "" {
		fmt.Fprintln(os.Stderr, "migrate: -path and -database are required")
		os.Exit(2)
	}

	m, err := migrate.New("file://"+*path, *database)
	if err != nil {
		log.Fatalf("migrate: failed to open: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	cmd := args[0]
	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up: %v", err)
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate down: %v", err)
		}
	case "steps":
		if len(args) < 2 {
			log.Fatalf("migrate steps: missing N")
		}
		n, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("migrate steps: invalid N %q: %v", args[1], err)
		}
		if err := m.Steps(n); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate steps: %v", err)
		}
	case "force":
		if len(args) < 2 {
			log.Fatalf("migrate force: missing V")
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("migrate force: invalid V %q: %v", args[1], err)
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("migrate force: %v", err)
		}
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("migrate version: %v", err)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
	case "create":
		// `migrate create -seq -ext sql NAME` → create 000NNN_NAME.up.sql + .down.sql
		// with zero-padded sequential numbering matching the dir's highest prefix.
		if len(args) < 2 {
			log.Fatalf("migrate create: missing NAME")
		}
		if err := createSeq(*path, args[1]); err != nil {
			log.Fatalf("migrate create: %v", err)
		}
	default:
		log.Fatalf("migrate: unknown command %q", cmd)
	}
}

// createSeq writes a new <NNN>_<name>.up.sql and .down.sql pair, numbering
// one past the highest existing prefix in dir. Mirrors the upstream
// `migrate create -seq -ext sql` behavior.
func createSeq(dir, name string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", dir, err)
	}
	maxSeq := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		prefix := strings.SplitN(e.Name(), "_", 2)[0]
		n, err := strconv.Atoi(prefix)
		if err != nil {
			continue
		}
		if n > maxSeq {
			maxSeq = n
		}
	}
	next := maxSeq + 1
	base := fmt.Sprintf("%06d_%s", next, name)
	up := filepath.Join(dir, base+".up.sql")
	down := filepath.Join(dir, base+".down.sql")
	if err := os.WriteFile(up, []byte("-- "+base+".up.sql\n"), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", up, err)
	}
	if err := os.WriteFile(down, []byte("-- "+base+".down.sql\n"), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", down, err)
	}
	fmt.Printf("created %s, %s\n", up, down)
	return nil
}
