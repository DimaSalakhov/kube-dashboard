package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.999-0700"})

	cfg := MustParseConfig()
	log.WithField("config", cfg).Info("Configuration loaded")

	kube, err := NewKube(cfg)
	if err != nil {
		log.Fatal(err)
	}

	releaseMonitor := NewReleaseMonitor(kube, cfg.Monitor)

	releaseMonitor.MustStart()

	h := MustNewHanlder(kube)

	log.Info("Up and running")

	router := mux.NewRouter()
	router.HandleFunc(`/`, h.getContexts).Methods("GET")
	router.HandleFunc(`/{context:[\w\-]+}`, h.contextDetails).Methods("GET")
	router.HandleFunc(`/{context:[\w\-]+}/{namespace:[\w\-]+}/deployments`, h.getDeployments).Methods("GET")

	http.ListenAndServe(":8080", router)
}

type handler struct {
	kube *kube
}

func MustNewHanlder(kube *kube) *handler {
	return &handler{
		kube: kube,
	}
}

func (h *handler) getContexts(w http.ResponseWriter, r *http.Request) {
	contexts := make([]struct{ Name string }, 0, len(h.kube.contexts))
	for context := range h.kube.contexts {
		contexts = append(contexts, struct{ Name string }{Name: context})
	}

	renderTemplate(w, "ui/templates/contexts.tmpl", []breadcrumb{},
		struct {
			Contexts []struct{ Name string }
		}{
			Contexts: contexts,
		})
}

func (h *handler) contextDetails(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	context := params["context"]

	ctx, ok := h.kube.contexts[context]
	if !ok {
		http.Error(w, fmt.Sprintf("Cannot find context [%s]", context), http.StatusNotFound)
		return
	}

	renderTemplate(w, "ui/templates/context.tmpl",
		[]breadcrumb{
			{Text: context},
		},
		struct {
			Context    string
			Namespaces []string
		}{
			Context:    context,
			Namespaces: ctx.namespaces,
		})
}

func (h *handler) getDeployments(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	context := params["context"]
	namespace := params["namespace"]

	ctx, ok := h.kube.contexts[context]
	if !ok {
		http.Error(w, fmt.Sprintf("Cannot find context [%s]", context), http.StatusNotFound)
		return
	}

	deployments, err := ctx.client.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "ui/templates/deployments.tmpl",
		[]breadcrumb{
			{Text: context, URL: fmt.Sprintf("/%s", context)},
			{Text: namespace},
			{Text: "Deployments"},
		},
		struct {
			Deployments []corev1.Deployment
		}{
			Deployments: deployments.Items,
		})
}
