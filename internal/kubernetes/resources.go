package kubernetes

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateSecretV2
func CreateSecretV2(inCluster string, secret *v1.Secret) error {
	_, clientset, _ := CreateKubeConfig(inCluster)

	_, err := clientset.CoreV1().Secrets(secret.Namespace).Create(
		context.Background(),
		secret,
		metav1.CreateOptions{},
	)
	if err != nil {
		return err
	}
	log.Infof("created Secret %s in Namespace %s\n", secret.Name, secret.Namespace)
	return nil
}

// ReadConfigMapV2
func ReadConfigMapV2(inCluster string, namespace string, configMapName string) (map[string]string, error) {
	_, clientset, _ := CreateKubeConfig(inCluster)

	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		return map[string]string{}, fmt.Errorf("error getting ConfigMap: %s", err)
	}

	parsedSecretData := make(map[string]string)
	for key, value := range configMap.Data {
		parsedSecretData[key] = string(value)
	}

	return parsedSecretData, nil
}

// ReadSecretV2
func ReadSecretV2(inCluster string, namespace string, secretName string) (map[string]string, error) {
	_, clientset, _ := CreateKubeConfig(inCluster)

	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return map[string]string{}, fmt.Errorf("error getting secret: %s", err)
	}

	parsedSecretData := make(map[string]string)
	for key, value := range secret.Data {
		parsedSecretData[key] = string(value)
	}

	return parsedSecretData, nil
}

// UpdateConfigMapV2
func UpdateConfigMapV2(inCluster string, namespace, configMapName string, key string, value string) error {
	_, clientset, _ := CreateKubeConfig(inCluster)

	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting ConfigMap: %s", err)
	}

	configMap.Data = map[string]string{key: value}
	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(
		context.Background(),
		configMap,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	log.Infof("updated ConfigMap %s in Namespace %s\n", configMap.Name, configMap.Namespace)

	return nil
}
