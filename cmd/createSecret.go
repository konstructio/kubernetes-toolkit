package cmd

import (
	"github.com/kubefirst/kubernetes-toolkit/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CreateK8sSecretCmdOptions *kubernetes.CreateK8sSecretCmdOptions = &kubernetes.CreateK8sSecretCmdOptions{}

// syncEcrTokenCmd represents the syncEcrToken command
var createK8sSecret = &cobra.Command{
	Use:   "create-k8s-secret",
	Short: "Create a Kubernetes secret if it does not exist ",
	Long:  `Create a Kubernetes secret if it does not exist `,
	Run: func(cmd *cobra.Command, args []string) {
		_, clientset, _ := kubernetes.CreateKubeConfig(CreateK8sSecretCmdOptions.KubeInClusterConfig)

		err := kubernetes.CreateK8sSecret(&clientset, CreateK8sSecretCmdOptions)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createK8sSecret)
	createK8sSecret.PersistentFlags().BoolVar(&CreateK8sSecretCmdOptions.KubeInClusterConfig, "use-kubeconfig-in-cluster", true, "Kube config type - in-cluster (default), set to false to use local")

	createK8sSecret.Flags().StringVar(&CreateK8sSecretCmdOptions.Namespace, "namespace", CreateK8sSecretCmdOptions.Namespace, "Kubernetes Namespace to create secret in (required)")
	err := createK8sSecret.MarkFlagRequired("namespace")
	if err != nil {
		log.Fatal(err)
	}
	createK8sSecret.Flags().StringVar(&CreateK8sSecretCmdOptions.Name, "name", CreateK8sSecretCmdOptions.Name, "secret name (required)")
	err = createK8sSecret.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err)
	}
}
