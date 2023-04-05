package kubernetes

import (
	"context"
	"fmt"
	"time"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	externalSecretsAPIVersion = "external-secrets.io/v1beta1"
)

// WaitForClusterSecretStoreReady
func WaitForClusterSecretStoreReady(restConfig *rest.Config, storeName string, timeoutSeconds int64) error {
	for i := int64(0); i <= timeoutSeconds; i++ {
		log.Infof("waiting for ClusterSecretStore %s", storeName)

		namespace := "external-secrets-operator"

		store := &esv1beta1.ClusterSecretStore{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
			},
			Spec: esv1beta1.ClusterSecretStore{
				Provider: &esv1beta1.SecretStoreProvider{
					Fake: &esv1beta1.FakeProvider{},
				},
			},
		}
		provider, err := esv1beta1.GetProvider(store)
		if err != nil {
			fmt.Println("error with eso provider")
		}

		kubeClient, err := client.New(restConfig, client.Options{})
		if err != nil {
			fmt.Println("error getting kube client")
		}
		
		esClientset, err := provider.NewClient(context.Background(), store, kubeClient, "external-secrets-operator")
		if err != nil {
			fmt.Println("error getting eso clientset")
		}

		es, err := esClientset.

		

		// // Call the API to return matched ClusterSecretStore objects
		// data, err := clientset.CoreV1().RESTClient().Get().
		// 	AbsPath(fmt.Sprintf("/apis/%s", externalSecretsAPIVersion)).
		// 	Resource("clustersecretstores").
		// 	Name(storeName).
		// 	DoRaw(context.Background())
		// if err != nil {
		// 	log.Info("error getting matched secret stores, checking again")
		// 	continue
		// }

		// // Unmarshal JSON API response to ClusterSecretStore object
		// var resp *esv1beta1.ClusterSecretStore
		// if err := json.Unmarshal(data, &resp); err != nil {
		// 	log.Errorf("error converting ClusterSecretStore data: %s", err)
		// }

		var lastCondition string
		for _, condition := range resp.Status.Conditions {
			switch {
			case condition.Type == esv1beta1.SecretStoreReady && condition.Status == v1.ConditionTrue:
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
