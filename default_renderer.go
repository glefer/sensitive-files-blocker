package sensitive_files_blocker

import (
	"net/http"
)

// Renderer is an interface that defines the Render method.
type Renderer interface {
	Render(rw http.ResponseWriter, r *http.Request) error
}

// DefaultRenderer is a struct that holds the body of the default renderer.
type DefaultRenderer struct {
	Body string
}

// Render writes a simple string to the response writer.
func (dr *DefaultRenderer) Render(rw http.ResponseWriter, _ *http.Request) error {
	http.Error(rw, dr.Body, http.StatusForbidden)
	return nil
}
