package sensitive_files_blocker

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp/syntax"
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
				Files: []string{"passwords.txt", "secrets.txt", ""},
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
		FileRegex: []string{
			"file.txt",
			"^begin",
			"end$",
			"^all$",
			"",
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

func TestNew_InvalidRegex(t *testing.T) {
	config := &Config{
		FileRegex: []string{
			"file.txt",
			"invalid[regex",
		},
	}

	_, err := New(context.TODO(), nil, config, "TestSensitiveFileBlocker")
	if err == nil {
		t.Error("Expected error, got nil")
	} else {
		var syntaxErr *syntax.Error
		if !errors.As(err, &syntaxErr) {
			t.Errorf("Expected error of type *regexp.SyntaxError, got %T", err)
		}
	}
}

func TestSensitiveFileBlocker_ExactFileMatchWithTemplateDisabled(t *testing.T) {
	config := &Config{
		Files: []string{"exactfile.txt"},
		Template: TemplateConfig{
			Enabled: false, // Disable template rendering
		},
	}

	sfb, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
	if err != nil {
		t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/exactfile.txt", nil)
	res := httptest.NewRecorder()

	sfb.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, res.Code)
	}
}

func TestSensitiveFileBlocker_CaseSensitiveFileBlocking(t *testing.T) {
	config := &Config{
		Files: []string{"SensitiveFile.txt"},
	}

	sfb, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
	if err != nil {
		t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sensitivefile.txt", nil)
	res := httptest.NewRecorder()

	sfb.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
	}
}

func TestSensitiveFileBlocker_NonMatchingRegexAndExactFile(t *testing.T) {
	config := &Config{
		Files: []string{"^start.*", "exactfile.txt"},
	}

	sfb, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
	if err != nil {
		t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/nonmatching.txt", nil)
	res := httptest.NewRecorder()

	sfb.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
	}
}

func TestSensitiveFileBlocker_MultipleRegexAndExactFileConflicts(t *testing.T) {
	config := &Config{
		Files: []string{"exactfile.txt", "^exact.*"},
	}

	sfb, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
	if err != nil {
		t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/exactfile.txt", nil)
	res := httptest.NewRecorder()

	sfb.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, res.Code)
	}
}

func TestSensitiveFileBlocker_EmptyFilePath(t *testing.T) {
	config := &Config{
		Files: []string{"somefile.txt"},
	}

	sfb, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
	if err != nil {
		t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	sfb.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
	}
}

type MockRendererWithRenderError struct{}

var errMockRenderer = errors.New("mock error")

func (m *MockRendererWithRenderError) Render(_ http.ResponseWriter, _ *http.Request) error {
	return errMockRenderer
}

func TestSensitiveFileBlocker_ServeHTTP_RenderError(t *testing.T) {
	testCases := []struct {
		name      string
		config    *Config
		requested string
	}{
		{
			name:      "Without Regex",
			config:    &Config{Files: []string{"testenv"}},
			requested: "/testenv",
		},
		{
			name:      "With Regex",
			config:    &Config{FileRegex: []string{"^testenv"}},
			requested: "/testenv2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sfb, err := New(context.TODO(), mockHandler(), tc.config, "TestSensitiveFileBlocker")
			if err != nil {
				t.Fatalf("Failed to create SensitiveFileBlocker: %v", err)
			}

			sfba, ok := sfb.(*SensitiveFileBlocker)
			if !ok {
				t.Fatalf("Expected SensitiveFileBlocker to be of type *SensitiveFileBlocker")
			}

			sfba.renderer = &MockRendererWithRenderError{}

			req := httptest.NewRequest(http.MethodGet, tc.requested, nil)
			res := httptest.NewRecorder()
			sfb.ServeHTTP(res, req)
			if http.StatusInternalServerError != res.Code {
				t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, res.Code)
			}
		})
	}
}

func TestNewFileLoggerErrors(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		wantErr bool
	}{
		{
			name:    "Expect error when logs are enabled",
			enabled: true,
			wantErr: true,
		},
		{
			name:    "Expect no error when logs are not enabled",
			enabled: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				FileRegex: []string{
					"file.txt",
					"^begin",
					"end$",
					"^all$",
					"",
				},
				Logs: LogsConfig{
					Enabled: tt.enabled,
					LogFile: "/nonexistent_directory/test_log",
				},
			}

			_, err := New(context.TODO(), mockHandler(), config, "TestSensitiveFileBlocker")
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
