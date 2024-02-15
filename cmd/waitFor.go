/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/kubefirst/kubernetes-toolkit/internal/kubernetes"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type WaitForCmdOptions struct {
	Namespace           string
	Name                string
	Label               string
	Timeout             int64
	KubeInClusterConfig string
}

var waitForCmdOptions *WaitForCmdOptions = &WaitForCmdOptions{}

// waitForCmd represents the waitFor command
var waitForCmd = &cobra.Command{
	Use:   "wait-for",
	Short: "Wait on something in Kubernetes to be ready",
	Long:  `Wait on a resource in Kubernetes to reach a ready state`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("waitFor called")
	},
}

// waitForDeploymentCmd represents the waitForDeploymentCmd command
var waitForDeploymentCmd = &cobra.Command{
	Use:   "deployment",
	Short: "Wait for a Deployment to be ready",
	Long:  `Wait for a Deployment to be ready`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(waitForCmdOptions)
		label := strings.Split(waitForCmdOptions.Label, "=")
		if len(label) != 2 {
			log.Fatalf("please check the provided label: %s", waitForCmdOptions.Label)
		}

		_, clientset, _ := kubernetes.CreateKubeConfig(waitForCmdOptions.KubeInClusterConfig)
		deployment, err := kubernetes.ReturnDeploymentObject(&clientset, label[0], label[1], waitForCmdOptions.Namespace, waitForCmdOptions.Timeout)
		if err != nil {
			log.Fatalf("error retrieving deployment object: %s", err)
		}
		_, err = kubernetes.WaitForDeploymentReady(&clientset, deployment, waitForCmdOptions.Timeout)
		if err != nil {
			log.Fatalf("error waiting for deployment object: %s", err)
		}
	},
}

// waitForPodCmd represents the waitForPodCmd command
var waitForPodCmd = &cobra.Command{
	Use:   "pod",
	Short: "Wait for a Pod to be ready",
	Long:  `Wait for a Pod to be ready`,
	Run: func(cmd *cobra.Command, args []string) {
		label := strings.Split(waitForCmdOptions.Label, "=")
		if len(label) != 2 {
			log.Fatalf("please check the provided label: %s", waitForCmdOptions.Label)
		}

		_, clientset, _ := kubernetes.CreateKubeConfig(waitForCmdOptions.KubeInClusterConfig)
		pod, err := kubernetes.ReturnPodObject(&clientset, label[0], label[1], waitForCmdOptions.Namespace, waitForCmdOptions.Timeout)
		if err != nil {
			log.Fatalf("error retrieving pod object: %s", err)
		}
		_, err = kubernetes.WaitForPodReady(&clientset, pod, waitForCmdOptions.Timeout)
		if err != nil {
			log.Fatalf("error waiting for pod object: %s", err)
		}
	},
}

// waitForStatefulSetCmd represents the waitForStatefulSetCmd command
var waitForStatefulSetCmd = &cobra.Command{
	Use:   "statefulset",
	Short: "Wait for a StatefulSet to be ready",
	Long:  `Wait for a StatefulSet to be ready`,
	Run: func(cmd *cobra.Command, args []string) {
		label := strings.Split(waitForCmdOptions.Label, "=")
		if len(label) != 2 {
			log.Fatalf("please check the provided label: %s", waitForCmdOptions.Label)
		}

		_, clientset, _ := kubernetes.CreateKubeConfig(waitForCmdOptions.KubeInClusterConfig)
		sts, err := kubernetes.ReturnStatefulSetObject(&clientset, label[0], label[1], waitForCmdOptions.Namespace, waitForCmdOptions.Timeout)
		if err != nil {
			log.Fatalf("error retrieving statefulset object: %s", err)
		}
		_, err = kubernetes.WaitForStatefulSetReady(&clientset, sts, waitForCmdOptions.Timeout, false)
		if err != nil {
			log.Fatalf("error waiting for statefulset object: %s", err)
		}
	},
}

// waitForClusterSecretStoreCmd represents the waitForClusterSecretStoreCmd command
var waitForClusterSecretStoreCmd = &cobra.Command{
	Use:   "cluster-secret-store",
	Short: "Wait for an External Secrets Operator cluster secret store to be ready",
	Long:  `Wait for an External Secrets Operator cluster secret store to be ready`,
	Run: func(cmd *cobra.Command, args []string) {
		_, clientset, _ := kubernetes.CreateKubeConfig(waitForCmdOptions.KubeInClusterConfig)
		err := kubernetes.WaitForClusterSecretStoreReady(&clientset, waitForCmdOptions.Name, waitForCmdOptions.Timeout)
		if err != nil {
			log.Fatalf("error waiting for ClusterSecretStore object: %s", err)
		}
	},
}

// waitForMinioBucketCmd represents the waitForMinioBucketCmd command
var waitForMinioBucketCmd = &cobra.Command{
	Use:   "minio-buckets",
	Short: "Wait for all minio buckets to be created",
	Long:  `Wait for all minio buckets to be created`,
	Run: func(cmd *cobra.Command, args []string) {
		minioEndpoint := "minio.minio.svc.cluster.local:9000"
		minioDefaultUsername := "k-ray"
		minioDefaultPassword := "feedkraystars"
		// Initialize minio client object.
		minioClient, err := minio.New(minioEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(minioDefaultUsername, minioDefaultPassword, ""),
			Secure: false,
			Region: "us-k3d-1",
		})
		if err != nil {
			log.Fatal("Error creating Minio client: %s", err)
		}

		buckets := []string{"chartmuseum", "argo-artifacts", "gitlab-backup", "kubefirst-state-store", "vault-backend"}

		// loop until all buckets exist
		for {
			allExist := true
			for _, bucket := range buckets {
				found, err := minioClient.BucketExists(context.Background(), bucket)
				if err != nil {
					log.Fatalln("Error checking bucket existence:", err)
				}
				if !found {
					allExist = false
					break
				}
			}
			if allExist {
				break
			}
			fmt.Println("waiting for all minio buckets to exist...")
			time.Sleep(5 * time.Second)
		}

		fmt.Println("all minio buckets created")
	},
}

// waitForVaultUnsealCmd represents the waitForVaultUnseal command
var waitForVaultUnsealCmd = &cobra.Command{
	Use:   "vault-unseal",
	Short: "Wait for vault to be unsealed",
	Long:  `Wait for vault to be unsealed`,
	Run: func(cmd *cobra.Command, args []string) {

		for {
			vaultRootTokenLookup, err := kubernetes.ReadSecretV2("true", "vault", "vault-unseal-secret")
			if err != nil {
				fmt.Println(err)
				time.Sleep(5 * time.Second)
			}
			if vaultRootTokenLookup["root-token"] != "" {
				break
			}
		}
		fmt.Println("vault successfully unsealed")
	},
}

// waitForVaultInitCompleteCmd represents the waitForVaultInitComplete command
var waitForVaultInitCompleteCmd = &cobra.Command{
	Use:   "vault-init-complete",
	Short: "Wait for vault to be configured with terraform",
	Long:  `Wait for vault to be configured with terraform`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := api.DefaultConfig()
		cfg.Address = "http://vault.vault.svc.cluster.local:8200"

		client, err := api.NewClient(cfg)
		if err != nil {
			log.Fatal(err)
		}

		for {
			// Read the secret from the Vault server
			secret, err := client.Logical().Read("secret/data/development/metaphor")
			if err == nil {
				// Check if the secret was found
				if secret != nil {
					break
				}
			}

			fmt.Println("Waiting for vault to terraform to apply, sleeping 5 seconds")
			time.Sleep(5 * time.Second)
		}
		fmt.Println("vault successfully hydrated")
	},
}

// waitForCertificateCmd represents the waitForCertificate command
var waitForCertificateCmd = &cobra.Command{
	Use:   "certificate",
	Short: "Wait for cert-manager Certificate creation",
	Long:  `Wait for cert-manager Certificate creation`,
	Run: func(cmd *cobra.Command, args []string) {
		restConfig, _, _ := kubernetes.CreateKubeConfig(waitForCmdOptions.KubeInClusterConfig)
		err := kubernetes.WaitForCertificateReady(restConfig, waitForCmdOptions.Namespace, waitForCmdOptions.Name, waitForCmdOptions.Timeout)
		if err != nil {
			log.Fatalf("error waiting for Certificate object: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(waitForCmd)
	waitForCmd.PersistentFlags().StringVar(&waitForCmdOptions.KubeInClusterConfig, "use-kubeconfig-in-cluster", "true", "Kube config type - in-cluster (default), set to false to use local")

	// waitForDeploymentCmd
	waitForCmd.AddCommand(waitForDeploymentCmd)
	waitForDeploymentCmd.Flags().StringVar(&waitForCmdOptions.Namespace, "namespace", waitForCmdOptions.Namespace, "Namespace containing the resource (required)")
	err := waitForDeploymentCmd.MarkFlagRequired("namespace")
	if err != nil {
		log.Fatal(err)
	}
	waitForDeploymentCmd.Flags().StringVar(&waitForCmdOptions.Label, "label", waitForCmdOptions.Label, "Label to select the resource in the form key=value (required)")
	err = waitForDeploymentCmd.MarkFlagRequired("label")
	if err != nil {
		log.Fatal(err)
	}
	waitForDeploymentCmd.Flags().Int64Var(&waitForCmdOptions.Timeout, "timeout-seconds", 60, "Timeout seconds - 60 (default)")

	// waitForMinioBucketCmd
	waitForCmd.AddCommand(waitForMinioBucketCmd)

	// waitForVaultUnsealCmd
	waitForCmd.AddCommand(waitForVaultUnsealCmd)

	// waitForVaultInitCompleteCmd
	waitForCmd.AddCommand(waitForVaultInitCompleteCmd)

	// waitForPodCmd
	waitForCmd.AddCommand(waitForPodCmd)
	waitForPodCmd.Flags().StringVar(&waitForCmdOptions.Namespace, "namespace", waitForCmdOptions.Namespace, "Namespace containing the resource (required)")
	err = waitForPodCmd.MarkFlagRequired("namespace")
	if err != nil {
		log.Fatal(err)
	}
	waitForPodCmd.Flags().StringVar(&waitForCmdOptions.Label, "label", waitForCmdOptions.Label, "Label to select the resource in the form key=value (required)")
	err = waitForPodCmd.MarkFlagRequired("label")
	if err != nil {
		log.Fatal(err)
	}
	waitForPodCmd.Flags().Int64Var(&waitForCmdOptions.Timeout, "timeout-seconds", 60, "Timeout seconds - 60 (default)")

	// waitForStatefulSetCmd
	waitForCmd.AddCommand(waitForStatefulSetCmd)
	waitForStatefulSetCmd.Flags().StringVar(&waitForCmdOptions.Namespace, "namespace", waitForCmdOptions.Namespace, "Namespace containing the resource (required)")
	err = waitForStatefulSetCmd.MarkFlagRequired("namespace")
	if err != nil {
		log.Fatal(err)
	}
	waitForStatefulSetCmd.Flags().StringVar(&waitForCmdOptions.Label, "label", waitForCmdOptions.Label, "Label to select the resource in the form key=value (required)")
	err = waitForStatefulSetCmd.MarkFlagRequired("label")
	if err != nil {
		log.Fatal(err)
	}
	waitForStatefulSetCmd.Flags().Int64Var(&waitForCmdOptions.Timeout, "timeout-seconds", 60, "Timeout seconds - 60 (default)")

	// waitForClusterSecretStoreCmd
	waitForCmd.AddCommand(waitForClusterSecretStoreCmd)
	waitForClusterSecretStoreCmd.Flags().StringVar(&waitForCmdOptions.Name, "name", waitForCmdOptions.Name, "Resource name (required)")
	err = waitForClusterSecretStoreCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err)
	}
	waitForClusterSecretStoreCmd.Flags().Int64Var(&waitForCmdOptions.Timeout, "timeout-seconds", 60, "Timeout seconds - 60 (default)")

	// waitForCertificateCmd
	waitForCmd.AddCommand(waitForCertificateCmd)
	waitForCertificateCmd.Flags().StringVar(&waitForCmdOptions.Namespace, "namespace", waitForCmdOptions.Namespace, "Namespace containing the resource (required)")
	err = waitForCertificateCmd.MarkFlagRequired("namespace")
	if err != nil {
		log.Fatal(err)
	}
	waitForCertificateCmd.Flags().StringVar(&waitForCmdOptions.Name, "name", waitForCmdOptions.Name, "Resource name (required)")
	err = waitForCertificateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err)
	}
	waitForCertificateCmd.Flags().Int64Var(&waitForCmdOptions.Timeout, "timeout-seconds", 60, "Timeout seconds - 60 (default)")
}
