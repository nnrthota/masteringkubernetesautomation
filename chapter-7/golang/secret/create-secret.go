package secret

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateSecret(
	clientset *kubernetes.Clientset,
	secret apiv1.Secret,
) error {
	secretClient := clientset.CoreV1().Secrets(apiv1.NamespaceDefault)
	result, err := secretClient.Create(&secret)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created Secret %q.\n", result.GetObjectMeta().GetName())
	return nil
}

func CreateWordpressSecret() apiv1.Secret {
	secret := apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mysql-pass",
		},
		Data: map[string][]byte{
			"password": []byte("yff37dqi893kdu"),
		},
	}
	return secret
}
