package main

import (
	"os"
	"path/filepath"

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
	config     *kubeconfig
	configPath string
}

func NewKube() (*kube, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to obtain user's home dir")
	}

	configPath := filepath.Join(dir, ".kube", "config")

	f, err := os.Open(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open kube config")
	}

	var config kubeconfig
	if err = yaml.NewDecoder(f).Decode(&config); err != nil {
		return nil, errors.Wrap(err, "Failed to parse kube config")
	}

	return &kube{
		config:     &config,
		configPath: configPath,
	}, nil
}

func (k *kube) getContexts() (map[string]context, error) {
	contexts := make(map[string]context)
	for _, c := range k.config.Contexts {
		client, err := k.buildClient(c.Name)
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

func (k *kube) buildClient(context string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: k.configPath},
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
