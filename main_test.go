package sensitive_files_blocker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSensitiveFileBlocker_SimpleFileName(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	config := &Config{
		Files: []string{"passwords.txt", "secrets.txt"},
	}

	sfb, err := New(context.TODO(), mockHandler, config, "TestSensitiveFileBlocker")
	if err != nil {
		t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/secrets.txt", nil)

	res := httptest.NewRecorder()

	sfb.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, res.Code)
	}
}

func TestSensitiveFileBlocker_RegexFileName(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	config := &Config{
		Files: []string{
			"file.txt",
			"^begin",
			"end$",
			"^all$",
		},
	}

	sfb, err := New(context.TODO(), mockHandler, config, "TestSensitiveFileBlocker")
	if err != nil {
		t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
	}

	testCases := []struct {
		fileName   string
		statusCode int
	}{
		{"all.txt", http.StatusOK},
		{"file.txt", http.StatusForbidden},
		{"begin.txt", http.StatusForbidden},
		{"end.txt", http.StatusOK},
		{"someend", http.StatusForbidden},
		{"all", http.StatusForbidden},
		{"notblocked.txt", http.StatusOK},
		{"beginandAnother", http.StatusForbidden},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/"+tc.fileName, nil)
		res := httptest.NewRecorder()
		sfb.ServeHTTP(res, req)
		if res.Code != tc.statusCode {
			t.Errorf("Expected status code %d for file %s, got %d", tc.statusCode, tc.fileName, res.Code)
		}
	}
}

func TestCreateConfig(t *testing.T) {
	config := CreateConfig()

	if config == nil {
		t.Errorf("CreateConfig returned nil")
		return
	}

	if len(config.Files) != 0 {
		t.Errorf("Expected Files to be an empty slice, got %v", config.Files)
	}
}

func TestNew_EmptyFiles(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	config := &Config{
		Files: []string{},
	}

	_, err := New(context.TODO(), mockHandler, config, "TestSensitiveFileBlocker")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
