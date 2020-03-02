package main

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	h := MustNewHanlder()

	log.Info("Up and running")

	router := mux.NewRouter()
	router.HandleFunc("/", h.getContexts).Methods("GET")
	router.HandleFunc("/{kubectx:\\w+}/deployments", h.getDeployments).Methods("GET")

	http.ListenAndServe(":8080", router)
}

type handler struct {
	kube *kube
}

func MustNewHanlder() *handler {
	kube, err := NewKube()
	if err != nil {
		log.Fatal(err)
	}

	return &handler{
		kube: kube,
	}
}

func (h *handler) getContexts(w http.ResponseWriter, r *http.Request) {
	contexts := make([]struct{ Name string }, 0, len(h.kube.config.Contexts))
	for _, v := range h.kube.config.Contexts {
		contexts = append(contexts, struct{ Name string }{Name: v.Name})
	}

	renderTemplate(w, "ui/templates/contexts.tmpl", []breadcrumb{},
		struct {
			Contexts []struct{ Name string }
		}{
			Contexts: contexts,
		})
}

func (h *handler) getDeployments(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	currentContext := params["kubectx"]

	namespace := h.kube.getNamespace(currentContext)

	client, err := h.kube.buildClient(currentContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	deployments, err := client.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "ui/templates/deployments.tmpl",
		[]breadcrumb{
			{Text: currentContext},
			{Text: "Deployments"},
		},
		struct {
			Deployments []corev1.Deployment
		}{
			Deployments: deployments.Items,
		})
}
