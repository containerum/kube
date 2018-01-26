package server

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// context keys
const (
	RequestObjectKey  = "requestObject"
	ResponseObjectKey = "responseObject"
	NamespaceKey      = "namespace"
	ObjectNameKey     = "objectName"
	KubeClientKey     = "kubeclient"
)

type KubeClient struct {
	Tag      string
	Client   *kubernetes.Clientset
	UseCount int
}

var KubeClients []KubeClient

// so far only one config supported
func LoadKubeClients(cfgpath string) (err error) {
	config, err := clientcmd.BuildConfigFromFlags("", cfgpath)
	if err != nil {
		logrus.Fatalf("failed to parse kube config: %v", err)
		return
	}
	kcli, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("failed to create kube client from %s: %v", cfgpath, err)
		return
	}

	KubeClients = append(KubeClients, KubeClient{
		Tag:      "kubeclient-1",
		Client:   kcli,
		UseCount: 0,
	})
	return
}
