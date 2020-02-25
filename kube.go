package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
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

func (k *kube) getNamespace(contextName string) string {
	for _, v := range k.config.Contexts {
		if strings.EqualFold(contextName, v.Name) {
			return v.Context.Namespace
		}
	}
	return ""
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
