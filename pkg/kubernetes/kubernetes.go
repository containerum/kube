package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//Kube is struct for kubernetes client
type Kube struct {
	*kubernetes.Clientset
	config *rest.Config
}

//RegisterClient creates kubernetes client
func (k *Kube) RegisterClient(cfgpath string) error {
	var config *rest.Config
	var err error

	if cfgpath == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", cfgpath)
	}
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
