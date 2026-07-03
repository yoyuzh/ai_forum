package bootstrap

import (
	"os"
	"strings"
	"testing"
)

func TestDockerComposeKeepsAPIServerInternalOnly(t *testing.T) {
	body, err := os.ReadFile("../../../docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	api := sliceService(text, "api-server:")
	if !strings.Contains(api, "expose:") || strings.Contains(api, "ports:") {
		t.Fatalf("api-server must use expose and no ports:\n%s", api)
	}
	worker := sliceService(text, "worker-service:")
	if !strings.Contains(worker, "- api-server") {
		t.Fatalf("worker-service must depend on api-server:\n%s", worker)
	}
}

func TestNginxBlocksInternalPath(t *testing.T) {
	body, err := os.ReadFile("../../../deploy/nginx.conf")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "location /internal/") || !strings.Contains(text, "return 404") {
		t.Fatalf("nginx must return 404 for /internal/:\n%s", text)
	}
}

func TestDockerfilePackagesRBACModel(t *testing.T) {
	body, err := os.ReadFile("../../Dockerfile")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "internal/rbac/model.conf") {
		t.Fatalf("Dockerfile must package RBAC model.conf:\n%s", text)
	}
}

func sliceService(text, start string) string {
	i := strings.Index(text, start)
	if i < 0 {
		return ""
	}
	rest := text[i:]
	lines := strings.Split(rest, "\n")
	out := []string{lines[0]}
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && strings.HasSuffix(strings.TrimSpace(line), ":") {
			break
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
