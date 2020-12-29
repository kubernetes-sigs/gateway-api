/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	gwclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	// ensure auth plugins are loaded
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	KubeClient *clientset.Clientset
	GWClient   *gwclientset.Clientset
)

// LoadClientset returns clientsets for connecting to kubernetes clusters.
func LoadClientset() (*clientset.Clientset, *gwclientset.Clientset, error) {
	config, err := kubeConfig()
	if err != nil {
		return nil, nil, err
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	saclient, err := gwclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return client, saclient, nil
}

// DynamicParams are used to call DynamicApply.
type DynamicParams struct {
	Path      string
	Namespace string
	Delete    bool
	Params    map[string]string
}

// DynamicApply creates or updates Kubernetes resources defined with YAML at the
// provided path. This supports references to directories or individual files.
func DynamicApply(dp DynamicParams) error {
	path, err := filepath.Abs(dp.Path)
	if err != nil {
		return fmt.Errorf("error calculating filepath: %w", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		files, ioErr := ioutil.ReadDir(path)
		if err != nil {
			return ioErr
		}
		for _, file := range files {
			applyErr := DynamicApply(DynamicParams{
				Path:      dp.Path + "/" + file.Name(),
				Namespace: dp.Namespace,
				Delete:    dp.Delete,
			})
			if applyErr != nil {
				return applyErr
			}
		}
		return nil
	}

	klog.V(5).Infof("Applying YAML in %s", path)
	config, err := kubeConfig()
	if err != nil {
		return fmt.Errorf("error getting kubeconfig: %w", err)
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading %s file: %w", path, err)
	}

	if len(dp.Params) > 0 {
		tmpl, tmplErr := template.New("dynamic").Parse(string(b))
		if tmplErr != nil {
			return fmt.Errorf("error templating %s file: %w", path, tmplErr)
		}
		out := bytes.Buffer{}
		err = tmpl.Execute(&out, dp.Params)
		if err != nil {
			return fmt.Errorf("error executing template for %s file: %w", path, err)
		}
		b = out.Bytes()
	}

	dd, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error initializing dynamic client: %w", err)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return fmt.Errorf("error initializing discovery client: %w", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(b), 4096)
	for {
		uObj := unstructured.Unstructured{}
		if err := decoder.Decode(&uObj); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			klog.Infof("manifest: %s", string(b))
			return fmt.Errorf("error parsing manifest: %w", err)
		}
		if len(uObj.Object) == 0 {
			// klog.Warningf("Found empty object in %s, continuing", path)
			continue
		}
		gvk := uObj.GroupVersionKind()
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if uObj.GetNamespace() == "" {
				uObj.SetNamespace(dp.Namespace)
			}
			dri = dd.Resource(mapping.Resource).Namespace(uObj.GetNamespace())
		} else {
			dri = dd.Resource(mapping.Resource)
		}

		if dp.Delete {
			delErr := dri.Delete(context.TODO(), uObj.GetName(), metav1.DeleteOptions{})
			if delErr != nil {
				if !apierrors.IsNotFound(delErr) {
					return fmt.Errorf("error deleting resource: %w", delErr)
				}
			}
			continue
		}

		res, err := dri.Get(context.TODO(), uObj.GetName(), metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("error getting resource: %w", err)
			}
			klog.V(5).Infof("Creating %s %s", uObj.GetName(), gvk.Kind)
			_, err = dri.Create(context.TODO(), &uObj, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("error creating resource: %w", err)
			}
			continue
		}
		uObj.SetResourceVersion(res.GetResourceVersion())
		klog.Infof("Updating %s %s", uObj.GetName(), gvk.Kind)
		_, err = dri.Update(context.TODO(), &uObj, metav1.UpdateOptions{})

		if err != nil {
			return fmt.Errorf("error updating resource: %w", err)
		}
	}

	return nil
}

// NewNamespace creates a new namespace using gateway-api-conformance- as
// prefix.
func NewNamespace() (string, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			// GenerateName: "gateway-api-conformance-",
			Name: "gateway-conformance",
			Labels: map[string]string{
				"app.kubernetes.io/name": "gateway-api-conformance",
			},
		},
	}

	var err error

	ns, err = KubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("unable to create namespace: %w", err)
	}

	return ns.Name, nil
}

// DeleteNamespace deletes a namespace and all the objects inside
func DeleteNamespace(namespace string) error {
	grace := int64(0)
	pb := metav1.DeletePropagationBackground

	return KubeClient.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
		PropagationPolicy:  &pb,
	})
}

// CleanupNamespaces removes namespaces created by conformance tests
func CleanupNamespaces() error {
	namespaces, err := KubeClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=gateway-api-conformance",
	})

	if err != nil {
		return err
	}

	for _, namespace := range namespaces.Items {
		err := DeleteNamespace(namespace.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewSelfSignedSecret creates a self signed SSL certificate and store it in a secret
func NewSelfSignedSecret(c clientset.Interface, namespace, secretName string, hosts []string) error {
	if len(hosts) == 0 {
		return fmt.Errorf("require a non-empty hosts for Subject Alternate Name values")
	}

	var serverKey, serverCert bytes.Buffer

	host := strings.Join(hosts, ",")

	if err := generateRSACert(host, &serverKey, &serverCert); err != nil {
		return err
	}

	data := map[string][]byte{
		corev1.TLSCertKey:       serverCert.Bytes(),
		corev1.TLSPrivateKeyKey: serverKey.Bytes(),
	}

	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: data,
	}

	if _, err := c.CoreV1().Secrets(namespace).Create(context.TODO(), newSecret, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

const (
	rsaBits  = 2048
	validFor = 365 * 24 * time.Hour
)

// generateRSACert generates a basic self signed certificate valir for a year
func generateRSACert(host string, keyOut, certOut io.Writer) error {
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "default",
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)

	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed creating cert: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return fmt.Errorf("failed creating key: %w", err)
	}

	return nil
}

func kubeConfig() (*restclient.Config, error) {
	config, err := restclient.InClusterConfig()
	if err != nil {
		// Attempt to use local KUBECONFIG
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
		// use the current context in kubeconfig
		var err error

		config, err = kubeconfig.ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	// TODO: add version information?
	config.UserAgent = fmt.Sprintf(
		"%s (%s/%s) gateway-api-conformance",
		filepath.Base(os.Args[0]),
		runtime.GOOS,
		runtime.GOARCH,
	)

	return config, nil
}
