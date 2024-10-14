package sensitive_files_blocker

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestFileLogger_Log(t *testing.T) {
	loggerInterface, _ := NewFileLogger(createTemporyFile(t))

	if _, ok := loggerInterface.(*FileLogger); !ok {
		t.Error("Expected NewFileLogger to return a FileLogger, but got a different type")
	}

	logger, ok := loggerInterface.(*FileLogger)
	if !ok {
		t.Fatal("Expected NewFileLogger to return a *FileLogger")
	}

	defer func() {
		err := logger.Close()
		if err != nil {
			t.Errorf("Failed to close file logger: %v", err)
		}
		err = os.Remove(logger.file.Name())
		if err != nil {
			t.Errorf("Failed to remove file: %v", err)
		}
	}()

	message := "test message"
	remoteAddr := "127.0.0.1:8080"
	expectedLog := remoteAddr[:len(remoteAddr)-6] + " " + message + "\n"

	logger.Log(message, remoteAddr)

	logBytes, err := os.ReadFile(logger.file.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(logBytes)
	if strings.HasSuffix(logContent, expectedLog) {
		t.Errorf("Expected log content:\n%s\nGot:\n%s", expectedLog, logContent)
	}
}

func TestFileLogger_Close(t *testing.T) {
	loggerInterface, _ := NewFileLogger(createTemporyFile(t))

	if _, ok := loggerInterface.(*FileLogger); !ok {
		t.Error("Expected NewFileLogger to return a FileLogger, but got a different type")
	}

	logger, ok := loggerInterface.(*FileLogger)
	if !ok {
		t.Fatal("Expected loggedInterface to return a *FileLogger")
	}

	err := logger.Close()
	if err != nil {
		t.Errorf("Failed to close file logger: %v", err)
	}

	err = logger.Close()
	if err == nil {
		t.Error("Expected error when closing already closed file logger, but got nil")
	}

	// Clean up the temporary file
	defer func() {
		err := os.Remove(logger.file.Name())
		if err != nil {
			t.Errorf("Failed to remove file: %v", err)
		}
	}()
}

func TestFileLoggerCreationAndLogging(t *testing.T) {
	loggerInterface, _ := NewFileLogger(createTemporyFile(t))

	if _, ok := loggerInterface.(*FileLogger); !ok {
		t.Error("Expected NewFileLogger to return a FileLogger, but got a different type")
	}

	logger, ok := loggerInterface.(*FileLogger)
	if !ok {
		t.Fatal("Expected loggedInterface to return a *FileLogger")
	}

	defer func() {
		err := logger.Close()
		if err != nil {
			t.Errorf("Failed to close file logger: %v", err)
		}
		err = os.Remove(logger.file.Name())
		if err != nil {
			t.Errorf("Failed to remove file: %v", err)
		}
	}()

	if logger.file == nil {
		t.Error("Expected file logger's file field to be non-nil, but got nil")
	}

	if logger.logger == nil {
		t.Error("Expected file logger's logger field to be non-nil, but got nil")
	}

	logger.Log("test message", "127.0.0.1:8080")

	// #nosec G304 This is a file logger write to tests
	logBytes, err := os.ReadFile(logger.file.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(logBytes)
	expectedLog := "127.0.0.1 test message"
	if strings.HasSuffix(logContent, expectedLog) {
		t.Errorf("Expected log content:\n%s\nGot:\n%s", expectedLog, logContent)
	}
}

func TestNewFileLogger_FileOpenError(t *testing.T) {
	filePath := "/nonexistent_directory/test_log"
	_, err := NewFileLogger(filePath)

	if err == nil {
		t.Error("Expected error when creating file logger with non-existent directory, but got nil")
	}
}

// mockCloser is a simple mock implementing io.Closer
type mockCloser struct {
	closed bool
	err    error
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.err
}

func (m *mockCloser) Log(_, _ string) {}

func TestSensitiveFileBlocker_Close_WithCloser(t *testing.T) {
	mock := &mockCloser{}
	var logger Logger = mock

	if _, ok := logger.(io.Closer); !ok {
		t.Log("Expected mockCloser to implement io.Closer, but it does not")
	}

	sfb := &SensitiveFileBlocker{
		logger: mock,
	}

	// Call Close
	err := sfb.Close()
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if !mock.closed {
		t.Error("Expected mockCloser to be closed, but it was not")
	}
}

func TestSensitiveFileBlocker_Close_With_NoopCloser(t *testing.T) {
	sfb := &SensitiveFileBlocker{
		logger: &NoopLogger{},
	}

	err := sfb.Close()
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

func createTemporyFile(t *testing.T) string {
	t.Helper()
	file, err := os.CreateTemp("", "test_log")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Failed to close file: %v", err)
	}
	return file.Name()
}
