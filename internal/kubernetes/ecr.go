package kubernetes

import (
	"context"
	"fmt"

	awsinternal "github.com/konstructio/kubernetes-toolkit/internal/aws"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SynchronizeECRTokenSecret
func SynchronizeECRTokenSecret(clientset *kubernetes.Clientset, o *SyncEcrCmdOptions) error {
	awsClient := &awsinternal.AWSConfiguration{
		Config: awsinternal.NewAwsV2(o.Region),
	}

	ecrToken, err := awsClient.GetECRAuthToken()
	if err != nil {
		return err
	}

	dockerConfigString := fmt.Sprintf(`{"auths": {"%s": {"auth": "%s"}}}`, o.RegistryURL, ecrToken)
	dockerCfgSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "docker-config", Namespace: o.Namespace},
		Data:       map[string][]byte{"config.json": []byte(dockerConfigString)},
		Type:       "Opaque",
	}

	log.Infof("using ecr registry url: %s", o.RegistryURL)

	// Determine if Secret already exists
	secretExists := true
	_, err = clientset.CoreV1().Secrets(dockerCfgSecret.ObjectMeta.Namespace).Get(context.Background(), dockerCfgSecret.Name, metav1.GetOptions{})
	if err != nil {
		log.Warn(err)
		secretExists = false
	}

	switch {
	// If the Secret already exists, update it
	case secretExists:
		log.Infof("secret %s/%s already exists, it will be updated", dockerCfgSecret.Namespace, dockerCfgSecret.Name)

		_, err = clientset.CoreV1().Secrets(dockerCfgSecret.ObjectMeta.Namespace).Update(context.TODO(), dockerCfgSecret, metav1.UpdateOptions{})
		if err != nil {
			log.Errorf("error creating kubernetes secret %s/%s: %s", dockerCfgSecret.Namespace, dockerCfgSecret.Name, err)
			return err
		}
		log.Infof("updated secret %s/%s with new ecr token", dockerCfgSecret.Namespace, dockerCfgSecret.Name)
	// If the Secret does not exist, create it
	case !secretExists:
		log.Infof("secret %s/%s does not exist, it will be created", dockerCfgSecret.Namespace, dockerCfgSecret.Name)

		_, err = clientset.CoreV1().Secrets(dockerCfgSecret.ObjectMeta.Namespace).Create(context.TODO(), dockerCfgSecret, metav1.CreateOptions{})
		if err != nil {
			log.Errorf("error creating kubernetes secret %s/%s: %s", dockerCfgSecret.Namespace, dockerCfgSecret.Name, err)
			return err
		}
		log.Infof("created secret %s/%s with new ecr token", dockerCfgSecret.Namespace, dockerCfgSecret.Name)
	}

	return nil
}
