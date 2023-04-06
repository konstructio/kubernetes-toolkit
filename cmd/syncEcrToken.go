package cmd

import (
	"github.com/kubefirst/kubernetes-toolkit/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var syncEcrCmdOptions *kubernetes.SyncEcrCmdOptions = &kubernetes.SyncEcrCmdOptions{}

// syncEcrTokenCmd represents the syncEcrToken command
var syncEcrTokenCmd = &cobra.Command{
	Use:   "sync-ecr-token",
	Short: "Retrieve a new ecr token and update an in-cluster secret containing the token",
	Long:  `Retrieve a new ecr token and update an in-cluster secret containing the token`,
	Run: func(cmd *cobra.Command, args []string) {
		_, clientset, _ := kubernetes.CreateKubeConfig(syncEcrCmdOptions.KubeInClusterConfig)

		err := kubernetes.SynchronizeECRTokenSecret(&clientset, syncEcrCmdOptions)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncEcrTokenCmd)
	syncEcrTokenCmd.PersistentFlags().BoolVar(&syncEcrCmdOptions.KubeInClusterConfig, "use-kubeconfig-in-cluster", true, "Kube config type - in-cluster (default), set to false to use local")

	syncEcrTokenCmd.Flags().StringVar(&syncEcrCmdOptions.Namespace, "namespace", syncEcrCmdOptions.Namespace, "Kubernetes Namespace to create/sync in (required)")
	err := syncEcrTokenCmd.MarkFlagRequired("namespace")
	if err != nil {
		log.Fatal(err)
	}
	syncEcrTokenCmd.Flags().StringVar(&syncEcrCmdOptions.Region, "region", syncEcrCmdOptions.Region, "AWS Region (required)")
	err = syncEcrTokenCmd.MarkFlagRequired("region")
	if err != nil {
		log.Fatal(err)
	}
	syncEcrTokenCmd.Flags().StringVar(&syncEcrCmdOptions.RegistryURL, "registry-url", syncEcrCmdOptions.RegistryURL, "ECR registry URL (required)")
	err = syncEcrTokenCmd.MarkFlagRequired("registry-url")
	if err != nil {
		log.Fatal(err)
	}
}
