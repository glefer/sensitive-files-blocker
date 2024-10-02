// Package sensitive_files_blocker is a Traefik plugin that blocks access to sensitive files based on the file name.
package sensitive_files_blocker

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

// Config struct holding configuration.
type Config struct {
	Files     []string       `json:"blockedFiles,omitempty"`
	FileRegex []string       `json:"blockedFilesRegex,omitempty"`
	Template  TemplateConfig `yaml:"template"`
}

// CreateConfig creates a new Config struct with default values.
func CreateConfig() *Config {
	return &Config{
		Files:     []string{},
		FileRegex: []string{},
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

// SensitiveFileBlocker holds the configuration and the next http.Handler.
type SensitiveFileBlocker struct {
	next       http.Handler
	filesExact map[string]struct{}
	filesRegex []*regexp.Regexp
	name       string
	renderer   Renderer
}

var errEmptyFileList = errors.New("files and FileRegex cannot be empty")

// New creates a new SensitiveFileBlocker with optimized matching.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Files) == 0 && len(config.FileRegex) == 0 {
		return nil, errEmptyFileList
	}

	filesExact := make(map[string]struct{})
	filesRegex := make([]*regexp.Regexp, 0, len(config.Files))

	for _, file := range config.Files {
		filesExact[file] = struct{}{}
	}

	for _, file := range config.FileRegex {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		re, err := regexp.Compile(file)
		if err != nil {
			return nil, err
		}
		filesRegex = append(filesRegex, re)
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
		filesExact: filesExact,
		filesRegex: filesRegex,
		next:       next,
		name:       name,
		renderer:   renderer,
	}, nil
}

// ServeHTTP checks for blocked files with optimized matching.
func (sfb *SensitiveFileBlocker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// Exact match check
	if _, found := sfb.filesExact[path]; found {
		err := sfb.renderer.Render(rw, req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Regex match check
	for _, re := range sfb.filesRegex {
		if re.MatchString(path) {
			err := sfb.renderer.Render(rw, req)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	sfb.next.ServeHTTP(rw, req)
}
