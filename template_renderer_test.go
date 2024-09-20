package sensitive_files_blocker

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

var errWrite = errors.New("write error")

// Mock ResponseWriter for testing
type mockResponseWriter struct {
	header      http.Header
	writtenData []byte
	writeErr    error
	statusCode  int
}

func (m *mockResponseWriter) Header() http.Header {
	if m.header == nil {
		m.header = http.Header{}
	}
	return m.header
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	m.writtenData = append(m.writtenData, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestNewTemplateRenderer(t *testing.T) {
	tests := []struct {
		name           string
		config         TemplateConfig
		expectedOutput string
		expectErr      bool
	}{
		{
			name: "Success",
			config: TemplateConfig{
				HTML: "<div>{{.Name}}</div>",
				Vars: map[string]interface{}{
					"Name": "John Doe",
				},
			},
			expectedOutput: "<div>John Doe</div>",
			expectErr:      false,
		},
		{
			name: "TemplateParseError",
			config: TemplateConfig{
				HTML: "<div>{{.Name}", // Missing closing }}
				Vars: map[string]interface{}{
					"Name": "John Doe",
				},
			},
			expectErr: true,
		},
		{
			name: "ExecuteError",
			config: TemplateConfig{
				HTML: `<div>{{call .Invalid}}</div>`,
				CSS:  "some-css",
				Vars: map[string]interface{}{
					"Invalid": "i am not a function",
				},
			},
			expectErr: true,
		},
		{
			name: "AddCSSVariable",
			config: TemplateConfig{
				HTML: "<style>{{.CSS}}</style><div>{{.Content}}</div>",
				Vars: map[string]interface{}{
					"Content": "Hello, World!",
				},
				CSS: "body { background-color: #fff; }",
			},
			expectedOutput: "<style>body { background-color: #fff; }</style><div>Hello, World!</div>",
			expectErr:      false,
		},
		{
			name: "NoCSS",
			config: TemplateConfig{
				HTML: "<style>{{.CSS}}</style><div>{{.Content}}</div>",
				Vars: map[string]interface{}{
					"Content": "Hello, World!",
				},
			},
			expectedOutput: "<style></style><div>Hello, World!</div>",
			expectErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, err := NewTemplateRenderer(tt.config)
			if (err != nil) != tt.expectErr {
				t.Fatalf("NewTemplateRenderer() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr {
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/", nil)

				err = renderer.Render(rec, req)
				if err != nil {
					t.Fatalf("unexpected error during rendering: %v", err)
				}

				if rec.Body.String() != tt.expectedOutput {
					t.Errorf("expected %q but got %q", tt.expectedOutput, rec.Body.String())
				}
			}
		})
	}
}

func TestTemplateRenderer_Render_WriteError(t *testing.T) {
	// Mock response writer with a simulated write error
	w := &mockResponseWriter{
		writeErr: errWrite,
	}

	// Instance of TemplateRenderer
	renderer := &TemplateRenderer{
		Body: "<html><body>Hello, world!</body></html>",
	}

	// Call the Render method
	err := renderer.Render(w, nil)

	if err == nil {
		t.Fatal("expected error but got none")
	}

	if !errors.Is(err, errWrite) {
		t.Fatalf("expected %q but got %q", errWrite, err)
	}
}
