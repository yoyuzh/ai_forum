package post

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostPackageDoesNotImportDownstreamDomains(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(".", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"/ai/", "/search", "/notification", "/moderation"} {
			if strings.Contains(string(body), forbidden) {
				t.Fatalf("%s imports forbidden domain %s", entry.Name(), forbidden)
			}
		}
	}
}
