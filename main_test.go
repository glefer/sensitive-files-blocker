package sensitive_files_blocker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestSensitiveFileBlocker(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		requestURL string
		wantStatus int
	}{
		{
			name: "SimpleFileName",
			config: &Config{
				Files: []string{"passwords.txt", "secrets.txt"},
			},
			requestURL: "/secrets.txt",
			wantStatus: http.StatusForbidden,
		},
		{
			name: "NonBlockedFile",
			config: &Config{
				Files: []string{"passwords.txt", "secrets.txt"},
			},
			requestURL: "/public.txt",
			wantStatus: http.StatusOK,
		},
		{
			name: "RegexFileName",
			config: &Config{
				Files: []string{"^regex.*"},
			},
			requestURL: "/regexfile.txt",
			wantStatus: http.StatusForbidden,
		},
		{
			name: "NonExistentFile",
			config: &Config{
				Files: []string{"passwords.txt", "secrets.txt"},
			},
			requestURL: "/nonexistent.txt",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfb, err := New(context.TODO(), mockHandler(), tt.config, "TestSensitiveFileBlocker")
			if err != nil {
				t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, tt.requestURL, nil)
			res := httptest.NewRecorder()

			sfb.ServeHTTP(res, req)
			if res.Code != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, res.Code)
			}
		})
	}
}

func TestSensitiveFileBlocker_RegexFileName(t *testing.T) {
	config := &Config{
		Files: []string{
			"file.txt",
			"^begin",
			"end$",
			"^all$",
		},
	}

	sfb, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
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
		t.Run(tc.fileName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tc.fileName, nil)
			res := httptest.NewRecorder()
			sfb.ServeHTTP(res, req)
			if res.Code != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, res.Code)
			}
		})
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
	config := &Config{
		Files: []string{},
	}

	_, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestNewWithTemplateEnabledReturnTemplateRender(t *testing.T) {
	config := CreateConfig()
	config.Files = []string{".env"}

	sfb, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
	if err != nil {
		t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
	}

	if sfb == nil {
		t.Fatalf("Expected SensitiveFileBlocker to not be nil")
	}

	sfba, ok := sfb.(*SensitiveFileBlocker)
	if !ok {
		t.Fatalf("Expected SensitiveFileBlocker to be of type *SensitiveFileBlocker")
	}

	if sfba.renderer == nil {
		t.Fatalf("Expected renderer to not be nil")
	}

	if _, ok := sfba.renderer.(*TemplateRenderer); !ok {
		t.Fatalf("Expected renderer to be of type *TemplateRenderer")
	}
}

// Test for invalid template syntax (parsing error).
func TestNewWithTemplateEnabledParseError(t *testing.T) {
	config := CreateConfig()
	config.Files = []string{".env"}
	config.Template.HTML = "<div>{{.Name}"

	_, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
	if err == nil {
		t.Fatal("expected template parsing error, but got none")
	}
}

type MockRenderer struct{}

func (t *MockRenderer) Render(_ http.ResponseWriter, _ *http.Request) error {
	return errEmptyFileList
}

func TestSensitiveFileBlocker_ServeHTTP_BlockFile(t *testing.T) {
	config := &Config{
		Files: []string{".env"},
	}

	sfb, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
	if err != nil {
		t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
	}

	sfba, ok := sfb.(*SensitiveFileBlocker)
	if !ok {
		t.Fatalf("Expected SensitiveFileBlocker to be of type *SensitiveFileBlocker")
	}

	sfba.renderer = &MockRenderer{}

	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	res := httptest.NewRecorder()
	sfb.ServeHTTP(res, req)
	if http.StatusInternalServerError != res.Code {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, res.Code)
	}
}
