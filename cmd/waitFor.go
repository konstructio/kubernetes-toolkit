/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kubefirst/kubernetes-toolkit/internal/kubernetes"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type WaitForCmdOptions struct {
	Namespace           string
	Label               string
	Timeout             int64
	KubeInClusterConfig bool
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

// waitForMinioBucketCmd represents the waitForMinioBucketCmd command
var waitForMinioBucketCmd = &cobra.Command{
	Use:   "minio-buckets",
	Short: "Wait for all minio buckets to be created",
	Long:  `Wait for all minio buckets to be created`,
	Run: func(cmd *cobra.Command, args []string) {

		minioPortForwardEndpoint := "minio.minio.svc.cluster.local:9000"
		minioDefaultUsername := "k-ray"
		minioDefaultPassword := "feedkraystars"
		// Initialize minio client object.
		minioClient, err := minio.New(minioPortForwardEndpoint, &minio.Options{
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

func init() {
	rootCmd.AddCommand(waitForCmd)
	waitForCmd.AddCommand(waitForDeploymentCmd)
	waitForCmd.AddCommand(waitForPodCmd)
	waitForCmd.AddCommand(waitForStatefulSetCmd)

	// Required flags
	var attach []*cobra.Command
	attach = append(attach, waitForDeploymentCmd, waitForPodCmd, waitForStatefulSetCmd)

	for _, command := range attach {
		command.Flags().StringVar(&waitForCmdOptions.Namespace, "namespace", waitForCmdOptions.Namespace, "Namespace containing the resource (required)")
		err := command.MarkFlagRequired("namespace")
		if err != nil {
			log.Fatal(err)
		}
		command.Flags().StringVar(&waitForCmdOptions.Label, "label", waitForCmdOptions.Label, "Label to select the resource in the form key=value (required)")
		err = command.MarkFlagRequired("label")
		if err != nil {
			log.Fatal(err)
		}
		command.Flags().Int64Var(&waitForCmdOptions.Timeout, "timeout-seconds", 60, "Timeout seconds - 60 (default)")
		command.Flags().BoolVar(&waitForCmdOptions.KubeInClusterConfig, "use-kubeconfig-in-cluster", true, "Kube config type - in-cluster (default), set to false to use local")
	}

	waitForMinioBucketCmd.Flags().BoolVar(&waitForCmdOptions.KubeInClusterConfig, "use-kubeconfig-in-cluster", true, "Kube config type - in-cluster (default), set to false to use local")
	waitForCmd.AddCommand(waitForMinioBucketCmd)
}
