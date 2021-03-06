package certs

import (
	"io/ioutil"
	"os"

	"github.com/dapr/dapr/pkg/credentials"
	"github.com/dapr/dapr/pkg/sentry/config"
	"github.com/dapr/dapr/pkg/sentry/kubernetes"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultSecretNamespace = "default"
)

// StoreCredentials saves the trust bundle in a Kubernetes secret store or locally on disk, depending on the hosting platform
func StoreCredentials(conf config.SentryConfig, rootCertPem, issuerCertPem, issuerKeyPem []byte) error {
	if config.IsKubernetesHosted() {
		return storeKubernetes(rootCertPem, issuerCertPem, issuerKeyPem)
	}
	return storeSelfhosted(rootCertPem, issuerCertPem, issuerKeyPem, conf.RootCertPath, conf.IssuerCertPath, conf.IssuerKeyPath)
}

func storeKubernetes(rootCertPem, issuerCertPem, issuerCertKey []byte) error {
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = defaultSecretNamespace
	}

	kubeClient, err := kubernetes.GetClient()
	if err != nil {
		return err
	}

	secret := &v1.Secret{
		Data: map[string][]byte{
			credentials.RootCertFilename:   rootCertPem,
			credentials.IssuerCertFilename: issuerCertPem,
			credentials.IssuerKeyFilename:  issuerCertKey,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeScrtName,
			Namespace: namespace,
		},
		Type: v1.SecretTypeOpaque,
	}

	// We update and not create because sentry expects a secret to already exist
	_, err = kubeClient.CoreV1().Secrets(namespace).Update(secret)
	if err != nil {
		return errors.Wrap(err, "failed saving secret to kubernetes")
	}
	return nil
}

/* #nosec */
func storeSelfhosted(rootCertPem, issuerCertPem, issuerKeyPem []byte, rootCertPath, issuerCertPath, issuerKeyPath string) error {
	err := ioutil.WriteFile(rootCertPath, rootCertPem, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed saving file to %s", rootCertPath)
	}

	err = ioutil.WriteFile(issuerCertPath, issuerCertPem, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed saving file to %s", issuerCertPath)
	}

	err = ioutil.WriteFile(issuerKeyPath, issuerKeyPem, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed saving file to %s", issuerKeyPath)
	}
	return nil
}
