package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// redactingCore wraps a zapcore.Core and masks any field whose key is in the
// redact set to the literal "***" before it reaches the underlying encoder.
// This is the enforcement point for "never log the full token" (§13.4).
//
// Known limitation (documented in P0 design risks): redaction is by exact field
// key name only. Nested/renamed token fields are not covered in P0; the
// internal-API middleware hardens this further in P3.
type redactingCore struct {
	inner  zapcore.Core
	redact map[string]struct{}
}

// newRedactingCore wraps inner so fields whose key is in redact are masked.
func newRedactingCore(inner zapcore.Core, redact map[string]struct{}) zapcore.Core {
	return &redactingCore{inner: inner, redact: redact}
}

func (c *redactingCore) Enabled(lvl zapcore.Level) bool { return c.inner.Enabled(lvl) }

func (c *redactingCore) With(fields []zap.Field) zapcore.Core {
	// Mask fields carried on the child core, then delegate.
	masked := make([]zap.Field, len(fields))
	for i, f := range fields {
		masked[i] = c.mask(f)
	}
	return &redactingCore{inner: c.inner.With(masked), redact: c.redact}
}

func (c *redactingCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	// Bind this core (not the inner one) so Write routes through our masking
	// Write method. Delegating to inner.Check would bind the entry to inner,
	// bypassing redaction.
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *redactingCore) Write(ent zapcore.Entry, fields []zap.Field) error {
	masked := make([]zap.Field, len(fields))
	for i, f := range fields {
		masked[i] = c.mask(f)
	}
	return c.inner.Write(ent, masked)
}

func (c *redactingCore) Sync() error { return c.inner.Sync() }

// mask returns a copy of f whose value is replaced with "***" when its key is
// in the redact set. Non-redacted fields pass through unchanged.
func (c *redactingCore) mask(f zap.Field) zap.Field {
	if _, ok := c.redact[f.Key]; !ok {
		return f
	}
	// Replace the value with the redaction literal, preserving the key and type
	// so the encoder still emits the field (just masked).
	return zap.Field{Key: f.Key, Type: zapcore.StringType, String: redactedValue}
}
