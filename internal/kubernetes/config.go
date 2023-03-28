package kubernetes

import (
	"os"
	"path/filepath"

	"github.com/kubefirst/kubernetes-toolkit/internal/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var fs afero.Fs = afero.NewOsFs()

// CreateKubeConfig
func CreateKubeConfig(inCluster bool) (*rest.Config, kubernetes.Clientset, string) {
	// inCluster is either true or false
	// If it's true, we pull Kubernetes API authentication from Pod SA
	// If it's false, we use local machine settings
	if inCluster {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		return config, *clientset, "in-cluster"
	}

	// Set path to kubeconfig
	kubeconfig := ReturnKubeConfigPath()

	// Check to make sure kubeconfig actually exists
	// If it doesn't, go fetch it
	if common.FileExists(fs, kubeconfig) {
		log.Debug("kubeconfig exists, moving on.")
	}

	// Show what path was set for kubeconfig
	log.Debugf("setting kubeconfig to: %s", kubeconfig)

	// Build configuration instance from the provided config file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("unable to locate kubeconfig file - checked path: %s", kubeconfig)
	}

	// Create clientset, which is used to run operations against the API
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return config, *clientset, kubeconfig
}

// ReturnKubeConfigPath generates the path in the filesystem to kubeconfig
func ReturnKubeConfigPath() string {
	var kubeconfig string
	// We expect kubeconfig to be available at ~/.kube/config
	// However, sometimes some people may use the env var $KUBECONFIG
	// to set the path to the active one - we will switch on that here
	if os.Getenv("KUBECONFIG") != "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	} else {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	return kubeconfig
}
