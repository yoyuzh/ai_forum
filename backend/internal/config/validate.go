package config

import (
	"fmt"
	"strings"
)

// looksLikeUnresolvedPlaceholder reports whether s is a literal "${VAR}"
// placeholder that Viper leaves in place when the referenced environment
// variable is unset. Such a value is NOT a real secret — accepting it would
// let startup succeed with placeholder strings standing in for secrets,
// defeating the fail-fast enforcement point (design D4).
func looksLikeUnresolvedPlaceholder(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 4 {
		return false
	}
	return strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}")
}

// Validate asserts that required secrets are present and resolved. JWT.Secret
// and InternalAPI.Token are always required. MySQL.Password is required unless
// Server.Mode == "debug" (local dev relaxation). A secret that is empty OR still
// a literal "${VAR}" placeholder is treated as missing. All missing keys are
// aggregated into a single error message so operators see every gap at once.
func Validate(cfg *Config) error {
	var missing []string

	if isMissingSecret(cfg.JWT.Secret) {
		missing = append(missing, "jwt.secret")
	}
	if isMissingSecret(cfg.InternalAPI.Token) {
		missing = append(missing, "internal_api.token")
	}
	// DB password is relaxed in debug mode for local development.
	if cfg.Server.Mode != "debug" && isMissingSecret(cfg.MySQL.Password) {
		missing = append(missing, "mysql.password")
	}
	// AI API key is required in non-debug mode — the AI client must fail at
	// startup, not at request time (design D4). Relaxed in debug for local dev.
	if cfg.Server.Mode != "debug" && isMissingSecret(cfg.AI.APIKey) {
		missing = append(missing, "ai.api_key")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config secrets: %s", strings.Join(missing, ", "))
	}
	return nil
}

// isMissingSecret reports whether a secret value is absent — either empty or
// an unresolved "${VAR}" placeholder left by Viper when the env var was unset.
func isMissingSecret(s string) bool {
	return s == "" || looksLikeUnresolvedPlaceholder(s)
}
