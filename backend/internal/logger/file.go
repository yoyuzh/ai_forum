package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	aiconfig "ai-forum/backend/internal/config"
)

// FileConfig controls optional lumberjack-based file rotation. When FilePath is
// empty, file rotation is disabled and logs go only to stderr.
type FileConfig struct {
	// FilePath is the log file path. Empty disables file output.
	FilePath string
	// MaxSizeMB is the maximum size in megabytes before rotation.
	MaxSizeMB int
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int
	// MaxAgeDays is the maximum number of days to retain old log files.
	MaxAgeDays int
}

// NewWithFile builds a logger that writes to stderr and, when FilePath is
// non-empty, a rotated log file via lumberjack. When FilePath is empty it
// behaves like New (stderr only). This satisfies the optional lumberjack
// plumbing (task 3.4); v1 ships stderr + optional file output.
func NewWithFile(cfg aiconfig.LogConfig, fc FileConfig) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}
	// Always include stderr so logs are never silently dropped.
	writers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stderr)}
	if fc.FilePath != "" {
		lj := &lumberjack.Logger{
			Filename:   fc.FilePath,
			MaxSize:    fc.MaxSizeMB,
			MaxBackups: fc.MaxBackups,
			MaxAge:     fc.MaxAgeDays,
		}
		writers = append(writers, zapcore.AddSync(lj))
	}
	combined := zapcore.NewMultiWriteSyncer(writers...)
	core := newCore(level, cfg.Encoding, combined)
	redact := newRedactSet(defaultRedactKeys)
	return &Logger{
		zap:    zap.New(newRedactingCore(core, redact)),
		redact: redact,
	}, nil
}
