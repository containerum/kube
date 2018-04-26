package kubernetes

import (
	"io"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
)

//TODO: Imp struct to GetPodLogs func
type LogOptions struct {
	Follow    bool
	Tail      int64
	Previous  bool
	Container string
}

//GetPodList returns pods list
func (k *Kube) GetPodList(ns string, owner string) (interface{}, error) {
	pods, err := k.CoreV1().Pods(ns).List(meta_v1.ListOptions{
		LabelSelector: getOwnerLabel(owner),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
			"Owner":     owner,
		}).Error(err)
		return nil, err
	}
	return pods, nil
}

//GetPod returns pod
func (k *Kube) GetPod(ns string, po string) (interface{}, error) {
	pod, err := k.CoreV1().Pods(ns).Get(po, meta_v1.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
			"Pod":       po,
		}).Error(err)
		return nil, err
	}
	return pod, nil
}

//DeletePod deletes pod
func (k *Kube) DeletePod(ns string, po string) error {
	err := k.CoreV1().Pods(ns).Delete(po, &meta_v1.DeleteOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"Namespace": ns,
			"Pod":       po,
		}).Error(err)
		return err
	}
	return nil
}

//GetPodLogs attaches client to pod log
func (k *Kube) GetPodLogs(ns string, po string, opt *LogOptions) (io.ReadCloser, error) {
	req := k.CoreV1().Pods(ns).GetLogs(po, &v1.PodLogOptions{
		TailLines: &opt.Tail,
		Follow:    opt.Follow,
		Previous:  opt.Previous,
		Container: opt.Container,
	})

	return req.Stream()
}
