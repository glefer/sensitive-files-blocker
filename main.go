// Package sensitive_files_blocker is a Traefik plugin that blocks access to sensitive files based on the file name.
package sensitive_files_blocker

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

// Config is a struct that holds the configuration for the SensitiveFileBlocker.
// It contains a slice of strings representing the names of the files to be blocked.
type Config struct {
	Files    []string       `json:"blockedFiles,omitempty"`
	Template TemplateConfig `yaml:"template"`
}

// CreateConfig creates a new Config struct with default values.
// It returns a pointer to the newly created Config struct.
func CreateConfig() *Config {
	return &Config{
		Files: []string{},
		Template: TemplateConfig{
			Enabled: true,
			HTML:    HTMLTemplate,
			CSS:     CSSTemplate,
			Vars: map[string]interface{}{
				"Title":   "403 Forbidden",
				"Heading": "403 Forbidden",
				"Body":    "You do not have permission to access this document.",
			},
		},
	}
}

// SensitiveFileBlocker is a struct that holds the configuration for the SensitiveFileBlocker.
// It contains a slice of strings representing the names of the files to be blocked.
// It also contains the next http.Handler in the chain.
// The name field is used to identify the middleware in the logs.
type SensitiveFileBlocker struct {
	next     http.Handler
	files    []string
	name     string
	renderer Renderer
}

var errEmptyFileList = errors.New("files list cannot be empty")

// New creates a new SensitiveFileBlocker with the given configuration.
// It returns a pointer to the newly created SensitiveFileBlocker and an error if the configuration is invalid.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Files) == 0 {
		return nil, errEmptyFileList
	}

	var renderer Renderer

	if config.Template.Enabled {
		var err error
		renderer, err = NewTemplateRenderer(config.Template)
		if err != nil {
			return nil, err
		}
	} else {
		renderer = &DefaultRenderer{
			Body: "You do not have permission to access this document.",
		}
	}

	return &SensitiveFileBlocker{
		files:    config.Files,
		next:     next,
		name:     name,
		renderer: renderer,
	}, nil
}

// ServeHTTP blocks access to files based on the file name.
// It checks if the file name matches any of the blocked files and returns a 403 Forbidden status code if it does.
func (sfb *SensitiveFileBlocker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, blockedFile := range sfb.files {
		if matched, _ := regexp.MatchString(blockedFile, strings.TrimLeft(req.URL.Path, "/")); matched {
			err := sfb.renderer.Render(rw, req)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	sfb.next.ServeHTTP(rw, req)
}
