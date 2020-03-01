package main

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func renderTemplate(w http.ResponseWriter, templatePath string, data interface{}) {
	templateName := filepath.Base(templatePath)

	partials := template.Must(template.ParseGlob("ui/templates/partials/*.tmpl"))
	_, err := partials.ParseFiles(templatePath)
	if err = partials.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
