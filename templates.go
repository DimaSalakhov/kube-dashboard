package main

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func renderTemplate(w http.ResponseWriter, templatePath string, breadcrumbs []breadcrumb, data interface{}) {
	templateName := filepath.Base(templatePath)

	partials := template.Must(template.ParseGlob("ui/templates/partials/*.tmpl"))
	_, err := partials.ParseFiles(templatePath)
	if err = partials.ExecuteTemplate(w, templateName, withBreadcrumbs(breadcrumbs, data)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func withBreadcrumbs(breadcrumbs []breadcrumb, data interface{}) interface{} {
	return struct {
		Result      interface{}
		Breadcrumbs []breadcrumb
	}{
		Result:      data,
		Breadcrumbs: breadcrumbs,
	}
}

type breadcrumb struct {
	URL  string
	Text string
}
