package kubernetes

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func randSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func random(seq int) string {
	rand.Seed(time.Now().UnixNano())
	return randSeq(seq)
}

// CreateK8sSecret
func CreateK8sSecret(clientset *kubernetes.Clientset, o *CreateK8sSecretCmdOptions) error {
	k1AccessToken := random(20)

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: o.Name, Namespace: o.Namespace},
		Data: map[string][]byte{
			"K1_ACCESS_TOKEN": []byte(k1AccessToken),
		},
	}

	_, err := clientset.CoreV1().Secrets(secret.ObjectMeta.Namespace).Get(context.TODO(), secret.ObjectMeta.Name, metav1.GetOptions{})
	if err == nil {
		fmt.Println("kubernetes secret %s/%s already created - skipping", secret.Namespace, secret.Name)
	} else if strings.Contains(err.Error(), "not found") {
		_, err = clientset.CoreV1().Secrets(secret.ObjectMeta.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("error creating kubernetes secret %s/%s: %s", secret.Namespace, secret.Name, err)
		}
		fmt.Println("created kubernetes secret: %s/%s", secret.Namespace, secret.Name)
	}
	return nil
}
