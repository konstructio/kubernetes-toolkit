package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

const (
	certManagerAPIVersion = "cert-manager.io/v1"
)

// WaitForCertificateReady
func WaitForCertificateReady(clientset *kubernetes.Clientset, namespace string, certificateName string, timeoutSeconds int64) error {
	for i := int64(0); i <= timeoutSeconds; i++ {
		log.Infof("waiting for Certificate %s", certificateName)

		// Call the API to return matched Certificate objects
		data, err := clientset.CoreV1().RESTClient().Get().
			AbsPath(fmt.Sprintf("/apis/%s", certManagerAPIVersion)).
			Namespace(namespace).
			Resource("certificates").
			Name(certificateName).
			DoRaw(context.Background())
		if err != nil {
			return fmt.Errorf("error retrieving Certificate: %s", err)
		}

		// Unmarshal JSON API response to Certificate object
		var resp *v1.Certificate
		if err := json.Unmarshal(data, &resp); err != nil {
			log.Errorf("error converting Certificate data: %s", err)
		}

		var lastCondition string
		for _, condition := range resp.Status.Conditions {
			switch {
			case condition.Type == v1.CertificateConditionReady && condition.Status == metav1.ConditionTrue:
				log.Infof("Certificate validated")
				return nil
			default:
				lastCondition = fmt.Sprintf("%s: %s", condition.Reason, condition.Message)
			}
		}

		if i == timeoutSeconds {
			return fmt.Errorf("timed out waiting for the Certificate to be ready: %s", lastCondition)
		}
		time.Sleep(time.Second * 1)
	}

	return nil
}
