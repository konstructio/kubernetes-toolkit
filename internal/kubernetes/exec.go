package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	terminal "golang.org/x/term"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ReadService reads a Kubernetes Service object
func ReadService(clientset *kubernetes.Clientset, namespace string, serviceName string) (*v1.Service, error) {
	service, err := clientset.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("error getting Service: %s", err)
		return &v1.Service{}, nil
	}

	return service, nil
}

// PodExecSession executes a command against a Pod
func PodExecSession(clientset *kubernetes.Clientset, config *rest.Config, p *PodSessionOptions, silent bool) error {
	// v1.PodExecOptions is passed to the rest client to form the req URL
	podExecOptions := v1.PodExecOptions{
		Stdin:   p.Stdin,
		Stdout:  p.Stdout,
		Stderr:  p.Stderr,
		TTY:     p.TtyEnabled,
		Command: p.Command,
	}

	err := podExec(clientset, config, p, podExecOptions, silent)
	if err != nil {
		return err
	}
	return nil
}

// podExec performs kube-exec on a Pod with a given command
func podExec(clientset *kubernetes.Clientset, config *rest.Config, ps *PodSessionOptions, pe v1.PodExecOptions, silent bool) error {
	// Format the request to be sent to the API
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(ps.PodName).
		Namespace(ps.Namespace).
		SubResource("exec")
	req.VersionedParams(&pe, scheme.ParameterCodec)

	// POST op against Kubernetes API to initiate remote command
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Fatalf("error executing command on Pod: %s", err)
		return err
	}

	// Put the terminal into raw mode to prevent it echoing characters twice
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		log.Fatalf("error when attempting to start terminal: %s", err)
		return err
	}
	defer terminal.Restore(0, oldState)

	var showOutput io.Writer
	if silent {
		showOutput = io.Discard
	} else {
		showOutput = os.Stdout
	}
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: showOutput,
		Stderr: os.Stderr,
		Tty:    ps.TtyEnabled,
	})
	if err != nil {
		log.Fatalf("error running command on Pod: %s", err)
	}
	return nil
}

// ReturnDeploymentObject returns a matching appsv1.Deployment object based on the filters
func ReturnDeploymentObject(clientset *kubernetes.Clientset, matchLabel string, matchLabelValue string, namespace string, timeoutSeconds int64) (*appsv1.Deployment, error) {
	// Filter
	deploymentListOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", matchLabel, matchLabelValue),
	}

	// Create watch operation
	objWatch, err := clientset.
		AppsV1().
		Deployments(namespace).
		Watch(context.Background(), deploymentListOptions)
	if err != nil {
		log.Fatalf("error when attempting to search for Deployment: %s", err)
	}
	log.Infof("waiting for %s Deployment to be created", matchLabelValue)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatalf("error waiting for %s Deployment to be created: %s", matchLabelValue, err)
			}
			if event.
				Object.(*appsv1.Deployment).Status.Replicas > 0 {
				spec, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), deploymentListOptions)
				if err != nil {
					log.Fatalf("error when searching for Deployment: %s", err)
					return nil, err
				}
				return &spec.Items[0], nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error("the Deployment was not created within the timeout period")
			return nil, fmt.Errorf("the Deployment was not created within the timeout period")
		}
	}
}

// ReturnPodObject returns a matching v1.Pod object based on the filters
func ReturnPodObject(clientset *kubernetes.Clientset, matchLabel string, matchLabelValue string, namespace string, timeoutSeconds int64) (*v1.Pod, error) {
	// Filter
	podListOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", matchLabel, matchLabelValue),
	}

	// Create watch operation
	objWatch, err := clientset.
		CoreV1().
		Pods(namespace).
		Watch(context.Background(), podListOptions)
	if err != nil {
		log.Fatalf("error when attempting to search for Pod: %s", err)
	}
	log.Infof("waiting for %s Pod to be created", matchLabelValue)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatalf("error waiting for %s Pod to be created: %s", matchLabelValue, err)
			}
			if event.
				Object.(*v1.Pod).Status.Phase == "Pending" {
				spec, err := clientset.CoreV1().Pods(namespace).List(context.Background(), podListOptions)
				if err != nil {
					log.Fatalf("error when searching for Pod: %s", err)
					return nil, err
				}
				return &spec.Items[0], nil
			}
			if event.
				Object.(*v1.Pod).Status.Phase == "Running" {
				spec, err := clientset.CoreV1().Pods(namespace).List(context.Background(), podListOptions)
				if err != nil {
					log.Fatalf("error when searching for Pod: %s", err)
					return nil, err
				}
				return &spec.Items[0], nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error("the Pod was not created within the timeout period")
			return nil, fmt.Errorf("the Pod was not created within the timeout period")
		}
	}
}

// ReturnStatefulSetObject returns a matching appsv1.StatefulSet object based on the filters
func ReturnStatefulSetObject(clientset *kubernetes.Clientset, matchLabel string, matchLabelValue string, namespace string, timeoutSeconds int64) (*appsv1.StatefulSet, error) {
	// Filter
	statefulSetListOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", matchLabel, matchLabelValue),
	}

	// Create watch operation
	objWatch, err := clientset.
		AppsV1().
		StatefulSets(namespace).
		Watch(context.Background(), statefulSetListOptions)
	if err != nil {
		log.Fatalf("error when attempting to search for StatefulSet: %s", err)
	}
	log.Infof("waiting for %s StatefulSet to be created using label %s=%s", matchLabelValue, matchLabel, matchLabelValue)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatalf("error waiting for %s StatefulSet to be created: %s", matchLabelValue, err)
			}
			if event.
				Object.(*appsv1.StatefulSet).Status.Replicas > 0 {
				spec, err := clientset.AppsV1().StatefulSets(namespace).List(context.Background(), statefulSetListOptions)
				if err != nil {
					log.Fatalf("error when searching for StatefulSet: %s", err)
					return nil, err
				}
				return &spec.Items[0], nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error("the StatefulSet was not created within the timeout period")
			return nil, fmt.Errorf("the StatefulSet was not created within the timeout period")
		}
	}
}

// WaitForDeploymentReady waits for a target Deployment to become ready
func WaitForDeploymentReady(clientset *kubernetes.Clientset, deployment *appsv1.Deployment, timeoutSeconds int64) (bool, error) {

	// Format list for metav1.ListOptions for watch
	configuredReplicas := deployment.Status.Replicas
	watchOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf(
			"metadata.name=%s", deployment.Name),
	}

	// Create watch operation
	objWatch, err := clientset.
		AppsV1().
		Deployments(deployment.ObjectMeta.Namespace).
		Watch(context.Background(), watchOptions)
	if err != nil {
		log.Fatalf("error when attempting to wait for Deployment: %s", err)
	}
	log.Infof("waiting for %s Deployment to be ready - this could take up to %v seconds", deployment.Name, timeoutSeconds)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatalf("error waiting for Deployment: %s", err)
			}
			if event.
				Object.(*appsv1.Deployment).
				Status.ReadyReplicas == configuredReplicas {
				log.Infof("all Pods in Deployment %s are ready", deployment.Name)
				return true, nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error("the Deployment was not ready within the timeout period")
			return false, fmt.Errorf("the Deployment was not ready within the timeout period")
		}
	}
}

// WaitForPodReady waits for a target Pod to become ready
func WaitForPodReady(clientset *kubernetes.Clientset, pod *v1.Pod, timeoutSeconds int64) (bool, error) {
	// Format list for metav1.ListOptions for watch
	watchOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf(
			"metadata.name=%s", pod.Name),
	}

	// Create watch operation
	objWatch, err := clientset.
		CoreV1().
		Pods(pod.ObjectMeta.Namespace).
		Watch(context.Background(), watchOptions)
	if err != nil {
		log.Fatalf("error when attempting to wait for Pod: %s", err)
	}
	log.Infof("waiting for %s Pod to be ready - this could take up to %v seconds", pod.Name, timeoutSeconds)

	// Feed events using provided channel
	objChan := objWatch.ResultChan()

	// Listen until the Pod is ready
	// Timeout if it isn't ready within timeoutSeconds
	for {
		select {
		case event, ok := <-objChan:
			if !ok {
				// Error if the channel closes
				log.Error("fail")
			}
			if event.
				Object.(*v1.Pod).
				Status.
				Phase == "Running" {
				log.Infof("Pod %s is %s", pod.Name, event.Object.(*v1.Pod).Status.Phase)
				return true, nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error("the operation timed out while waiting for the Pod to become ready")
			return false, fmt.Errorf("the operation timed out while waiting for the Pod to become ready")
		}
	}
}

// WaitForStatefulSetReady waits for a target StatefulSet to become ready
func WaitForStatefulSetReady(clientset *kubernetes.Clientset, statefulset *appsv1.StatefulSet, timeoutSeconds int64, ignoreReady bool) (bool, error) {

	// Format list for metav1.ListOptions for watch
	configuredReplicas := statefulset.Status.Replicas

	// Create watch operation
	objWatch, err := clientset.AppsV1().StatefulSets(statefulset.ObjectMeta.Namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf(
			"metadata.name=%s", statefulset.Name),
	})
	if err != nil {
		log.Fatalf("error when attempting to wait for StatefulSet: %s", err)
	}
	log.Infof("waiting for %s StatefulSet to be ready - this could take up to %v seconds", statefulset.Name, timeoutSeconds)

	objChan := objWatch.ResultChan()
	for {
		select {
		case event, ok := <-objChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatalf("error waiting for StatefulSet: %s", err)
			}
			if ignoreReady {
				// Under circumstances where Pods may be running but not ready
				// These may require additional setup before use, etc.
				currentRevision := event.Object.(*appsv1.StatefulSet).Status.CurrentRevision
				if event.Object.(*appsv1.StatefulSet).Status.CurrentReplicas == configuredReplicas {
					// Get Pods owned by the StatefulSet
					pods, err := clientset.CoreV1().Pods(statefulset.ObjectMeta.Namespace).List(context.Background(), metav1.ListOptions{
						LabelSelector: fmt.Sprintf("controller-revision-hash=%s", currentRevision),
					})
					if err != nil {
						log.Fatalf("could not find Pods owned by StatefulSet")
					}

					// Determine when the Pods are running
					for _, pod := range pods.Items {
						err := watchForStatefulSetPodReady(clientset, statefulset.Namespace, statefulset.Name, pod.Name, timeoutSeconds)
						if err != nil {
							log.Fatalf(err.Error())
						}
						log.Infof("pod %s in statefulset %s is running", pod.Name, statefulset.Name)
					}
					objWatch.Stop()
					return true, nil
				}
			} else {
				// Under normal circumstances, once all Pods are ready
				// return success
				if event.Object.(*appsv1.StatefulSet).Status.AvailableReplicas == configuredReplicas {
					log.Infof("all Pods in StatefulSet %s are ready", statefulset.Name)
					objWatch.Stop()
					return true, nil
				}
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error("the StatefulSet was not ready within the timeout period")
			return false, fmt.Errorf("the StatefulSet was not ready within the timeout period")
		}
	}
}

// watchForStatefulSetPodReady inspects a Pod associated with a StatefulSet and
// uses a channel to determine when it's ready
// The channel will timeout if the Pod isn't ready by timeoutSeconds
func watchForStatefulSetPodReady(clientset *kubernetes.Clientset, namespace string, statefulSetName string, podName string, timeoutSeconds int64) error {
	podObjWatch, err := clientset.CoreV1().Pods(namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf(
			"metadata.name=%s", podName),
	})
	if err != nil {
		log.Fatalf("error when attempting to wait for Pod: %s", err)
	}

	podObjChan := podObjWatch.ResultChan()
	for {
		select {
		case podEvent, ok := <-podObjChan:
			time.Sleep(time.Second * 1)
			if !ok {
				// Error if the channel closes
				log.Fatalf("error waiting for Pod: %s", err)
			}
			if podEvent.Object.(*v1.Pod).Status.Phase == "Running" {
				podObjWatch.Stop()
				return nil
			}
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			log.Error("the StatefulSet Pod was not ready within the timeout period")
			return errors.New("the StatefulSet Pod was not ready within the timeout period")
		}
	}
}
