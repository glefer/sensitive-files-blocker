package sensitive_files_blocker

import (
	"io"
	"log"
	"os"
)

const ownerReadWrite = 0o600

// LogsConfig defines the logger traefik configuration.
type LogsConfig struct {
	Enabled bool   `json:"enabled"`
	LogFile string `json:"logFile,omitempty"`
}

// Logger interface for logging.
// The io.Closer interface is needed for compatibility with yaegi.
type Logger interface {
	io.Closer
	Log(message, remoteAddr string)
}

// NoopLogger does not perform any logging.
type NoopLogger struct{}

// Log does not perform any logging.
func (l *NoopLogger) Log(_, _ string) {}

// Close does not perform any logging.
func (l *NoopLogger) Close() error { return nil }

// FileLogger containing file for I/O and a logger to log directly.
type FileLogger struct {
	file   *os.File
	logger *log.Logger
}

// NewFileLogger creates a new FileLogger to the specified file.
// nolint: ireturn
func NewFileLogger(filePath string) (Logger, error) {
	// #nosec G304 This is a file logger, so it's expected to be configured by the user installing the plugin.
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, ownerReadWrite)
	if err != nil {
		return nil, err
	}

	return &FileLogger{
		file:   file,
		logger: log.New(file, "", log.LstdFlags),
	}, nil
}

// Log writes the message to the log file.
func (l *FileLogger) Log(message, remoteAddr string) {
	remoteAddr = remoteAddr[:len(remoteAddr)-5]

	l.logger.Printf("%s %s\n", remoteAddr, message)
}

// Close closes the log file (io.Closer interface).
func (l *FileLogger) Close() error {
	return l.file.Close()
}
