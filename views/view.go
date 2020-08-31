package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// LayoutDir is the directory to layouts
// TemplateExt is the extension name for the files
var (
	LayoutDir   = "views/layouts/"
	TemplateExt = ".gohtml"
)

// NewView automates the logic for views
func NewView(layout string, files ...string) *View {
	files = append(
		files,
		layoutFiles()...,
	)

	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

// Returns all layout files as a slice of strings
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}
	return files
}

// View is a data structure for html templates
type View struct {
	Template *template.Template
	Layout   string
}

// Render renders a view
func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

// ServeHTTP ensures views.View implements http.Handler
// which in turn is taken by mux.Router.Handle()
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := v.Render(w, nil); err != nil {
		panic(err)
	}
}
