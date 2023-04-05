package kubernetes

import (
	"context"
	"fmt"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	cl "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// WaitForCertificateReady
func WaitForCertificateReady(restConfig *rest.Config, namespace string, certificateName string, timeoutSeconds int64) error {
	for i := int64(0); i <= timeoutSeconds; i++ {
		log.Infof("waiting for Certificate %s", certificateName)

		cmclient, err := cl.NewForConfig(restConfig)
		if err != nil {
			return err
		}

		cert, err := cmclient.CertmanagerV1().Certificates(namespace).Get(context.Background(), certificateName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		var lastCondition string
		for _, condition := range cert.Status.Conditions {
			switch {
			case condition.Type == certmanagerv1.CertificateConditionReady && condition.Status == certmanagermetav1.ConditionTrue:
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
