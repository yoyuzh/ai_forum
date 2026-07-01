// Package logger provides structured zap logging with secret redaction and
// contextual field helpers, per architecture §13.1/§13.4.
package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	aiconfig "ai-forum/backend/internal/config"
)

// redactedValue is the literal every redacted field is replaced with.
const redactedValue = "***"

// defaultRedactKeys are the secret field names redacted on every logger by
// default, so the "INTERNAL_API_TOKEN must never appear in logs" guarantee
// holds even when a caller forgets WithRedaction (architecture §13.4).
//
// Beyond the bare names, this includes the common aliases a developer might
// use for the project's most sensitive secrets: "internal_api_token" (the
// viper config key is internal_api.token) and "authorization" (HTTP header).
// Redaction is exact-key only (see redact.go limitation note); nested fields
// under zap.Object are NOT covered in P0.
var defaultRedactKeys = []string{
	"token", "password", "secret", "api_key",
	"internal_api_token", "authorization",
}

// Logger is a thin wrapper around *zap.Logger that supports contextual child
// loggers and secret redaction. Fields bound via With live on the underlying
// core (not a separate logger-level list), so redaction composes correctly
// with With regardless of call order.
type Logger struct {
	zap    *zap.Logger
	redact map[string]struct{}
}

// New builds a zap logger from cfg writing to os.Stderr. JSON encoder when
// cfg.Encoding == "json", console encoder otherwise. Level comes from
// cfg.Level. Redaction is enabled by default for defaultRedactKeys.
func New(cfg aiconfig.LogConfig) (*Logger, error) {
	return NewWithWriter(cfg, os.Stderr)
}

// newCore builds the zapcore.Core for the given encoder/level, writing to w.
func newCore(level zapcore.Level, encoding string, w zapcore.WriteSyncer) zapcore.Core {
	var enc zapcore.Encoder
	switch encoding {
	case "json":
		enc = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	default:
		enc = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	return zapcore.NewCore(enc, w, level)
}

// NewWithWriter builds a logger writing to w. Used by tests and by the file
// rotation path (lumberjack). Redaction is enabled by default for
// defaultRedactKeys.
func NewWithWriter(cfg aiconfig.LogConfig, w io.Writer) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}
	core := newCore(level, cfg.Encoding, zapcore.AddSync(w))
	redact := newRedactSet(defaultRedactKeys)
	return &Logger{
		zap:    zap.New(newRedactingCore(core, redact)),
		redact: redact,
	}, nil
}

// newRedactSet builds the redact lookup set from a key list.
func newRedactSet(keys []string) map[string]struct{} {
	m := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		m[k] = struct{}{}
	}
	return m
}

// parseLevel maps a config level string to a zapcore.Level. Unknown levels
// default to Info.
func parseLevel(s string) (zapcore.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info", "":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return 0, fmt.Errorf("unknown log level %q", s)
	}
}

// With returns a child logger that binds the given contextual fields to every
// subsequent entry. Fields are bound at the core level so they survive any
// later WithRedaction call (and are themselves masked if their key is in the
// redact set).
func (l *Logger) With(fields ...zap.Field) *Logger {
	child := l.clone()
	child.zap = zap.New(l.zap.Core().With(fields))
	return child
}

// WithRedaction returns a logger whose core masks any field whose name is in
// keys to the literal "***", in addition to the default redact keys. Because
// fields bound via With live on the core, this preserves earlier With-bound
// fields. This is the enforcement point for "never log the full token"
// (architecture §13.4); the default keys already cover the common case.
func (l *Logger) WithRedaction(keys ...string) *Logger {
	merged := make(map[string]struct{}, len(l.redact)+len(keys))
	for k := range l.redact {
		merged[k] = struct{}{}
	}
	for _, k := range keys {
		merged[k] = struct{}{}
	}
	child := l.clone()
	child.redact = merged
	child.zap = zap.New(newRedactingCore(l.zap.Core(), merged))
	return child
}

// clone copies the logger without sharing the zap instance (so With/WithRedaction
// do not mutate the receiver).
func (l *Logger) clone() *Logger {
	return &Logger{
		zap:    l.zap,
		redact: l.redact,
	}
}

// --- Logging methods (delegate to the underlying zap logger) ---

// Debug logs at Debug level.
func (l *Logger) Debug(msg string, fields ...zap.Field) { l.zap.Debug(msg, fields...) }

// Info logs at Info level.
func (l *Logger) Info(msg string, fields ...zap.Field) { l.zap.Info(msg, fields...) }

// Warn logs at Warn level.
func (l *Logger) Warn(msg string, fields ...zap.Field) { l.zap.Warn(msg, fields...) }

// Error logs at Error level.
func (l *Logger) Error(msg string, fields ...zap.Field) { l.zap.Error(msg, fields...) }

// Sync flushes buffered log output. It is safe to call at shutdown. Errors
// from syncing non-file sinks (e.g. stderr, which cannot fsync) are ignored,
// matching zap's own convention — only real file-sync failures surface.
func (l *Logger) Sync() error {
	if err := l.zap.Sync(); err != nil {
		// zap returns "sync /dev/stderr: bad file descriptor" for stderr; that
		// is expected and not actionable, so swallow it.
		if !isBenignSyncError(err) {
			return err
		}
	}
	return nil
}

// isBenignSyncError reports whether a Sync error comes from a non-file sink
// (stderr/stdout) that cannot be fsynced and is therefore expected. The match
// is scoped to the sink PATH (/dev/stderr, /dev/stdout) only — a bare
// "bad file descriptor" is NOT matched, because syscall.EBADF from a real
// file sink (lumberjack fd invalidated by disk full / fd exhaustion / storage
// unmount) produces identical text and must propagate so callers know logs
// were lost.
func isBenignSyncError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "/dev/stderr") || strings.Contains(msg, "/dev/stdout")
}

// --- Contextual field helpers (architecture §13.1) ---

// FieldEventID binds an event identifier.
func FieldEventID(v string) zap.Field { return zap.String("event_id", v) }

// FieldTaskID binds an Asynq task identifier.
func FieldTaskID(v string) zap.Field { return zap.String("task_id", v) }

// FieldUserID binds the acting user identifier.
func FieldUserID(v int64) zap.Field { return zap.Int64("user_id", v) }

// FieldRequestID binds a request/correlation identifier.
func FieldRequestID(v string) zap.Field { return zap.String("request_id", v) }

// FieldPostID binds a post identifier.
func FieldPostID(v int64) zap.Field { return zap.Int64("post_id", v) }

// FieldCommentID binds a comment identifier.
func FieldCommentID(v int64) zap.Field { return zap.Int64("comment_id", v) }

// FieldAIAgentID binds an AI agent identifier.
func FieldAIAgentID(v string) zap.Field { return zap.String("ai_agent_id", v) }

// FieldTriggerType binds the AI trigger kind (e.g. mention, followup, manual).
func FieldTriggerType(v string) zap.Field { return zap.String("trigger_type", v) }

// FieldToken binds a token value. It is masked to "***" by the default
// redaction set on every logger.
func FieldToken(v string) zap.Field { return zap.String("token", v) }

// FieldInternalAPI binds the worker→api-server internal API token. Its key
// "internal_api_token" is in the default redact set, so the value is always
// masked — the full INTERNAL_API_TOKEN must never appear in logs (§13.4).
func FieldInternalAPI(v string) zap.Field { return zap.String("internal_api_token", v) }
