package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type releaseMonitor struct {
	kube        *kube
	contexts    []string
	deployments map[string]deployment
}

type deployment struct {
	image string
}

func NewReleaseMonitor(kube *kube, config monitorConfig) *releaseMonitor {
	return &releaseMonitor{
		kube:        kube,
		contexts:    config.Contexts,
		deployments: make(map[string]deployment),
	}
}

func (m *releaseMonitor) MustStart() {
	go func() {
		notify := false
		for {
			log.Debug("Polling deployments")

			for ctxName, context := range m.kube.contexts {
				deployments, err := context.client.AppsV1().Deployments("").List(metav1.ListOptions{})
				if err != nil {
					log.Error(errors.Wrapf(err, "Failed to get deployments for context [%s]", ctxName))
				}

				for _, d := range deployments.Items {
					image := d.Spec.Template.Spec.Containers[0].Image
					key := fmt.Sprintf("%s/%s/%s", ctxName, d.Namespace, d.Name)
					current, ok := m.deployments[key]
					if !ok || strings.EqualFold(image, current.image) == false {
						m.deployments[key] = deployment{image: image}
						if notify {
							notifyDeploymetChange(d.Name, image, ctxName)
						}
					}
				}
			}

			notify = true
			time.Sleep(30 * time.Second)
		}
	}()
}

func notifyDeploymetChange(name string, image string, context string) {
	log.Infof("Deployment [%s] with image [%s] was released to [%s]", name, image, context)
}
