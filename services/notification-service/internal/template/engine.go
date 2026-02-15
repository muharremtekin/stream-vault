package template

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

// Engine wraps Go's html/template with embedded templates.
type Engine struct {
	templates *template.Template
}

// New creates a new template Engine by parsing all embedded HTML templates.
func New() (*Engine, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	return &Engine{templates: tmpl}, nil
}

// Render executes the named template with the given data and returns the HTML string.
func (e *Engine) Render(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("executing template %q: %w", name, err)
	}
	return buf.String(), nil
}
