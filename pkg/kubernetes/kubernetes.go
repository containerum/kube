package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Kube struct {
	*kubernetes.Clientset
}

func (k *Kube) RegisterClient(cfgpath string) error {
	config, err := clientcmd.BuildConfigFromFlags("", cfgpath)
	if err != nil {
		panic(err)
	}
	kubecli, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	k.Clientset = kubecli
	return nil
}

func getOwnerLabel(owner string) (label string) {
	if owner != "" {
		label = fmt.Sprintf("owner=%s", owner)
	}
	return
}
