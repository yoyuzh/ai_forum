package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var tablePattern = regexp.MustCompile(`(?is)\b(create|drop)\s+table(?:\s+if\s+(?:not\s+)?exists)?\s+` + "`?" + `([a-zA-Z0-9_]+)` + "`?")

func main() {
	dir := "migrations"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := check(dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func check(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	owners := map[string][]string{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".sql") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		for _, match := range tablePattern.FindAllStringSubmatch(string(data), -1) {
			key := strings.ToLower(match[1] + " " + match[2])
			owners[key] = append(owners[key], name)
		}
	}
	var failures []string
	for key, files := range owners {
		if len(files) > 1 {
			sort.Strings(files)
			failures = append(failures, fmt.Sprintf("%s: %s", key, strings.Join(files, ", ")))
		}
	}
	sort.Strings(failures)
	if len(failures) > 0 {
		return fmt.Errorf("duplicate migration table ownership:\n%s", strings.Join(failures, "\n"))
	}
	return nil
}
