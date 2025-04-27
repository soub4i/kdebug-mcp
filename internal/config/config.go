package config

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func SwitchContext(ctx string) (*kubernetes.Clientset, error) {

	kubeconfigPath := getKubeconfig()

	cfg, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: ctx,
		}).ClientConfig()

	return kubernetes.NewForConfig(cfg)

}

func BuildConfig() (*kubernetes.Clientset, error) {

	kubeconfigPath := getKubeconfig()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	return kubernetes.NewForConfig(config)
}

func getKubeconfig() string {
	kubeconfigPath := os.Getenv("KUBECONFIG")

	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	return kubeconfigPath
}
