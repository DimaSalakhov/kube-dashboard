package main

import (
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
)

func renderTemplate(w http.ResponseWriter, templatePath string, breadcrumbs []breadcrumb, data interface{}) {
	templateName := filepath.Base(templatePath)

	partials := template.Must(template.New("base").Funcs(tempalteFunctions).ParseGlob("ui/templates/partials/*.tmpl"))
	_, err := partials.ParseFiles(templatePath)
	if err = partials.ExecuteTemplate(w, templateName, withBreadcrumbs(breadcrumbs, data)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var tempalteFunctions = template.FuncMap{
	"formatContainerImage": formatContainerImage,
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

var awsECRImageRegex = regexp.MustCompile(`.+\.dkr\.ecr\..*\.amazonaws.com\/(.+)`)

func formatContainerImage(image string) template.HTML {
	result := image

	if matches := awsECRImageRegex.FindStringSubmatch(image); matches != nil {
		result = matches[1]
	}

	return template.HTML(result)
}
