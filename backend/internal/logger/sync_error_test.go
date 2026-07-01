package logger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestIsBenignSyncError_StderrPathSwallowed asserts the /dev/stderr and
// /dev/stdout sync errors (which cannot fsync) are treated as benign.
func TestIsBenignSyncError_StderrPathSwallowed(t *testing.T) {
	assert.True(t, isBenignSyncError(errors.New("sync /dev/stderr: bad file descriptor")))
	assert.True(t, isBenignSyncError(errors.New("sync /dev/stdout: bad file descriptor")))
}

// TestIsBenignSyncError_FileEBADFNotSwallowed asserts a "bad file descriptor"
// error that does NOT come from stderr/stdout is NOT swallowed. This is the
// review HIGH finding: the bare "bad file descriptor" substring match silently
// ate real lumberjack file-sync failures (disk full, fd exhaustion, storage
// unmount), causing buffered log entries to be lost without the caller knowing.
//
// syscall.EBADF.Error() == "bad file descriptor" — identical text to the
// stderr case minus the path prefix. The filter must scope on the sink path,
// not on the generic EBADF text.
func TestIsBenignSyncError_FileEBADFNotSwallowed(t *testing.T) {
	// A bare EBADF with no /dev/stderr or /dev/stdout prefix — what a lumberjack
	// file sink returns when its fd is invalidated.
	err := errors.New("bad file descriptor")
	assert.False(t, isBenignSyncError(err),
		"bare 'bad file descriptor' from a file sink must NOT be swallowed — only /dev/stderr|stdout paths are benign")
}

// TestIsBenignSyncError_RealFileErrorsNotSwallowed asserts genuine file I/O
// failures propagate.
func TestIsBenignSyncError_RealFileErrorsNotSwallowed(t *testing.T) {
	assert.False(t, isBenignSyncError(errors.New("no space left on device")))
	assert.False(t, isBenignSyncError(errors.New("permission denied")))
}

// TestSync_FileSinkErrorPropagates is a behavioral test: when the underlying
// core's Sync returns a non-benign error, Logger.Sync must surface it rather
// than returning nil. We build a logger over a failingWriteSyncer that returns
// a real file error on Sync (simulating a lumberjack sink on a full disk).
func TestSync_FileSinkErrorPropagates(t *testing.T) {
	core := newCore(zapcore.InfoLevel, "json", &failingWriteSyncer{err: errors.New("no space left on device")})
	l := &Logger{zap: zap.New(core), redact: newRedactSet(defaultRedactKeys)}

	if err := l.Sync(); err == nil {
		t.Fatal("Logger.Sync must propagate non-benign file errors, got nil")
	}
}

// failingWriteSyncer is a zapcore.WriteSyncer whose Sync returns a configured
// non-benign error, simulating a lumberjack file sink on a full disk.
type failingWriteSyncer struct{ err error }

func (f *failingWriteSyncer) Write(p []byte) (int, error) { return len(p), nil }
func (f *failingWriteSyncer) Sync() error                 { return f.err }
