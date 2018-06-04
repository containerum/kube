package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//Kube is struct for kubernetes client
type Kube struct {
	*kubernetes.Clientset
	config *rest.Config
}

//RegisterClient creates kubernetes client
func (k *Kube) RegisterClient(cfgpath string) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	kubecli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	k.Clientset = kubecli
	k.config = config
	return nil
}

func getOwnerLabel(owner string) (label string) {
	if owner != "" {
		label = fmt.Sprintf("owner=%s", owner)
	}
	return
}
