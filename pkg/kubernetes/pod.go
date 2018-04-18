package kubernetes

import (
	"bytes"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
)

//TODO: Imp struct to GetPodLogs func
type LogOptions struct {
	Follow     bool
	StopFollow chan struct{}
	Tail       int64
	Previous   bool
	Container  string
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
func (k *Kube) GetPodLogs(ns string, po string, out *bytes.Buffer, opt *LogOptions) error {
	defer log.Debug("STOP FOLLOW LOGS STREAM")
	req := k.CoreV1().Pods(ns).GetLogs(po, &v1.PodLogOptions{
		TailLines: &opt.Tail,
		Follow:    opt.Follow,
		Previous:  opt.Previous,
		Container: opt.Container,
	})
	readCloser, err := req.Stream()
	if err != nil {
		log.WithError(err).Debug("STREAM")
		return err
	}
	defer readCloser.Close()
	for {
		select {
		case <-opt.StopFollow:
			log.WithError(err).Debug("FOLLOW")
			return nil
		default:
			buf := make([]byte, 1024)
			_, err := readCloser.Read(buf)
			if err != nil {
				log.WithError(err).Debug("READ")
				return err
			}
			_, err = out.Write(buf)
			if err != nil {
				log.WithError(err).Debug("WRITE")
				return err
			}
		}
	}
}
