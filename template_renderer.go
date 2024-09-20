package sensitive_files_blocker

import (
	"bytes"
	"html/template"
	"net/http"
)

// TemplateConfig defines the template configuration.
type TemplateConfig struct {
	Enabled bool   `yaml:"enabled"`
	HTML    string `yaml:"html"`
	CSS     string `yaml:"css"`
	Vars    map[string]interface{}
}

// TemplateRenderer used to display the forbidden page with HTML template.
type TemplateRenderer struct {
	Body string
}

// NewTemplateRenderer creates a new TemplateRenderer with the given configuration.
// nolint: ireturn
func NewTemplateRenderer(templateConfig TemplateConfig) (Renderer, error) {
	tmpl, err := template.New("forbidden").Parse(templateConfig.HTML)
	if err != nil {
		return nil, err
	}

	if _, ok := templateConfig.Vars["CSS"]; !ok {
		templateConfig.Vars["CSS"] = template.CSS(templateConfig.CSS) // #nosec G203 This template is configured by the user installing plugin.
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, templateConfig.Vars); err != nil {
		return nil, err
	}

	return &TemplateRenderer{
		Body: tpl.String(),
	}, nil
}

// Render writes the template to the response writer.
func (t *TemplateRenderer) Render(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusForbidden)
	_, err := w.Write([]byte(t.Body))
	if err != nil {
		return err
	}
	return nil
}

// HTMLTemplate defines the default HTML template used by default if template enabled.
const HTMLTemplate = `<!DOCTYPE html>
<html>
	<head>
		<title>{{ .Title }}</title>
		<style>{{ .CSS }}</style>
	</head>
	<body>

    <div class="container">
        <h1>{{ .Heading }}</h1>
        <p>{{ .Body }}</p>
        <a href="/">Go Back Home</a>
    </div>
	</body>
</html>
`

// CSSTemplate defines the CSS template used by default if template enabled.
const CSSTemplate = `
body {
    font-family: 'Arial', sans-serif;
    background-color: #f2f2f2;
    color: #333;
    margin: 0;
    padding: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh;
}
.container {
    text-align: center;
}
h1 {
    font-size: 6em;
    margin: 0;
    color: #55a2f4;
}
p {
    font-size: 1.5em;
    margin: 20px 0;
}
a {
    text-decoration: none;
    color: #55a2f4;
    border: 2px solid #55a2f4;
    padding: 10px 20px;
    border-radius: 5px;
    transition: 0.3s;
}
a:hover {
    background-color: #55a2f4;
    color: white;
}
`
