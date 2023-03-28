/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/kubefirst/kubernetes-toolkit/internal/kubernetes"
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
}
