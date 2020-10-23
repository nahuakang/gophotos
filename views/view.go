package views

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

// LayoutDir is the directory to layouts
// TemplateExt is the extension name for the files
var (
	LayoutDir   = "views/layouts/"
	TemplateDir = "views/"
	TemplateExt = ".gohtml"
)

// addTemplatePath takes in a slice of strings
// representing file paths for templates and prepends
// the TemplateDir directory to each string in the slice
//
// E.g. the input {"home"} would result in the output
// {"views/home"} if TemplateDir == "views/"
func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// addTemplateExt takes in a slice of strings
// representing file paths for templates and appends
// the TemplateExt extension to each string in the slice
//
// E.g. the input {"home"} would result in the output
// {"home.gohtml"} if TemplateExt == ".gohtml"
func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}

// NewView automates the logic for views
func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)

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
func (v *View) Render(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "text/html")

	switch data.(type) {
	case Data:
		// do nothing
	default:
		// Convert data to the Data struct type
		data = Data{
			Yield: data,
		}
	}

	var buf bytes.Buffer
	err := v.Template.ExecuteTemplate(&buf, v.Layout, data)
	if err != nil {
		http.Error(w, "Something went wrong. If the problem "+
			"persists, please email support@gophotos.com",
			http.StatusInternalServerError)
		return
	}

	io.Copy(w, &buf)
}

// ServeHTTP ensures views.View implements http.Handler
// which in turn is taken by mux.Router.Handle()
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, nil)
}
