package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//Kube is struct for kubernetes client
type Kube struct {
	*kubernetes.Clientset
}

//RegisterClient creates kubernetes client
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
