package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	externalSecretsAPIVersion = "external-secrets.io/v1beta1"
)

// WaitForClusterSecretStoreReady
func WaitForClusterSecretStoreReady(clientset *kubernetes.Clientset, storeName string, timeoutSeconds int64) error {
	for i := int64(0); i <= timeoutSeconds; i++ {
		log.Infof("waiting for ClusterSecretStore %s", storeName)

		// Call the API to return matched ClusterSecretStore objects
		data, err := clientset.CoreV1().RESTClient().Get().
			AbsPath(fmt.Sprintf("/apis/%s", externalSecretsAPIVersion)).
			Resource("clustersecretstores").
			Name(storeName).
			DoRaw(context.Background())
		if err != nil {
			log.Info("error getting matched secret stores, checking again")
		}

		// Unmarshal JSON API response to ClusterSecretStore object
		var resp *v1beta1.ClusterSecretStore
		if err := json.Unmarshal(data, &resp); err != nil {
			log.Errorf("error converting ClusterSecretStore data: %s", err)
		}

		var lastCondition string
		for _, condition := range resp.Status.Conditions {
			switch {
			case condition.Type == v1beta1.SecretStoreReady && condition.Status == v1.ConditionTrue:
				log.Infof("ClusterSecretStore validated")
				return nil
			default:
				lastCondition = fmt.Sprintf("%s: %s", condition.Reason, condition.Message)
			}
		}

		if i == timeoutSeconds {
			return fmt.Errorf("timed out waiting for the ClusterSecretStore to be ready: %s", lastCondition)
		}
		time.Sleep(time.Second * 1)
	}

	return nil
}
