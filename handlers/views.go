package handlers

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"sync"
)

// ViewResolver manages template rendering and caching.
type ViewResolver struct {
	dir           string
	templateCache templates
}

func NewViewResolver(dir string) *ViewResolver {
	return &ViewResolver{dir: dir, templateCache: *newTemplates()}
}

// Dir returs the root directory of ViewResolver.
func (r *ViewResolver) Dir() string {
	return r.dir
}

// ExecuteView render a view template with the provided model and
// writes it to the given writer.
func (r *ViewResolver) ExecuteView(name string, w io.Writer, model any) error {
	tmpl, err := r.FindView(name)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, name, model)
}

// FindView retrieves and parses a template by its name. It uses caching.
func (r *ViewResolver) FindView(name string) (*template.Template, error) {
	r.templateCache.RLock()
	tmpl, ok := r.templateCache.templates[name]
	r.templateCache.RUnlock()

	if ok {
		return tmpl, nil
	}

	r.templateCache.Lock()
	defer r.templateCache.Unlock()

	// Double-check to prevent race conditions.
	if tmpl, ok := r.templateCache.templates[name]; ok {
		return tmpl, nil
	}

	path := r.GetViewPath(name)
	tmpl, err := template.New(name).ParseFiles(path)
	if err != nil {
		return nil, err
	}

	r.templateCache.templates[name] = tmpl
	return tmpl, nil
}

// GetViewPath returns the absolute path to a view file, ensuring safety.
func (r *ViewResolver) GetViewPath(name string) string {
	safeName := filepath.Clean(name)
	return fmt.Sprintf("%s/%s.html", r.dir, safeName)
}

// templates provides a thread-safe map for caching templates.
type templates struct {
	sync.RWMutex
	templates map[string]*template.Template
}

func newTemplates() *templates {
	return &templates{templates: make(map[string]*template.Template)}
}
