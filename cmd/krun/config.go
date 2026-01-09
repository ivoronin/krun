package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeConfig struct {
	restConfig *rest.Config
	namespace  string
	clientset  *kubernetes.Clientset
}

func NewKubeConfig() (*KubeConfig, error) {
	config, namespace, err := getConfigAndNamespace()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &KubeConfig{
		restConfig: config,
		namespace:  namespace,
		clientset:  clientset,
	}, nil
}

func (k *KubeConfig) RESTConfig() *rest.Config {
	return k.restConfig
}

func (k *KubeConfig) Namespace() string {
	return k.namespace
}

func (k *KubeConfig) Clientset() *kubernetes.Clientset {
	return k.clientset
}

func getConfigAndNamespace() (*rest.Config, string, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		namespace, err := getCurrentInClusterNamespace()
		if err != nil {
			return nil, "", fmt.Errorf("error getting in-cluster namespace: %w", err)
		}

		return config, namespace, nil
	}

	if !errors.Is(err, rest.ErrNotInCluster) {
		return nil, "", fmt.Errorf("error checking in-cluster config: %w", err)
	}

	// Fall back to kubeconfig
	return getConfigAndNamespaceWithPath(clientcmd.RecommendedHomeFile)
}

func getConfigAndNamespaceWithPath(kubeconfigPath string) (*rest.Config, string, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = kubeconfigPath

	overrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, "", fmt.Errorf("error getting namespace: %w", err)
	}

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("error getting client config: %w", err)
	}

	return config, namespace, nil
}

func getCurrentInClusterNamespace() (string, error) {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", fmt.Errorf("failed to read namespace from service account: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}
