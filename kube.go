package main

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type kubeconfig struct {
	Contexts []struct {
		Name    string
		Context struct {
			Namespace string
		}
	}
}

type context struct {
	client     *kubernetes.Clientset
	namespaces []string
}

type kube struct {
	contexts map[string]context
}

func NewKube(cfg config) (*kube, error) {
	f, err := os.Open(cfg.KubeconfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open kube config")
	}

	var config kubeconfig
	if err = yaml.NewDecoder(f).Decode(&config); err != nil {
		return nil, errors.Wrap(err, "Failed to parse kube config")
	}

	contexts, err := buildContexts(cfg.KubeconfigPath, &config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to build contexts")
	}

	return &kube{
		contexts: contexts,
	}, nil
}

func buildContexts(configPath string, config *kubeconfig) (map[string]context, error) {
	contexts := make(map[string]context)
	for _, c := range config.Contexts {
		client, err := buildClient(configPath, c.Name)
		if err != nil {
			return nil, err
		}

		namespaceList, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		namespaces := make([]string, 0, len(namespaceList.Items))
		for _, ns := range namespaceList.Items {
			namespaces = append(namespaces, ns.Name)
		}

		contexts[c.Name] = context{
			client:     client,
			namespaces: namespaces,
		}
	}

	return contexts, nil
}

func buildClient(configPath string, context string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: configPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to build kube client config")
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to build kube client")
	}

	return client, err
}
