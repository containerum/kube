package server

import (
	"k8s.io/client-go/kubernetes"
)

type KubeClient struct {
	Tag      string
	Client   *kubernetes.Clientset
	UseCount int
}

var KubeClients []KubeClient
