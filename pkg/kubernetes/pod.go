package kubernetes

import (
	"bytes"
	"errors"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
)

var (
	tailDefault = int64(100)
)

var (
	ErrUnableGetPodList = errors.New("Unable to get pod list")
	ErrUnableGetPod     = errors.New("Unable to get pod")
)

//TODO: Imp struct to GetPodLogs func
type LogOptions struct {
	Follow     bool
	StopFollow chan struct{}
	Tail       int64
	Previous   bool
	Container  string
}

func (k *Kube) GetPodList(ns string, owner string) (interface{}, error) {
	pods, err := k.CoreV1().Pods(ns).List(meta_v1.ListOptions{
		LabelSelector: getOwnerLabel(owner),
	})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ns,
			"Owner":     owner,
		}).Error(ErrUnableGetPodList)
		return nil, ErrUnableGetPodList
	}
	return pods, nil
}

func (k *Kube) GetPod(ns string, po string) (interface{}, error) {
	pod, err := k.CoreV1().Pods(ns).Get(po, meta_v1.GetOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Namespace": ns,
			"Pod":       po,
		}).Error(ErrUnableGetPod)
		return nil, ErrUnableGetPod
	}
	return pod, nil
}

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
		return err
	}
	defer readCloser.Close()
	for {
		select {
		case <-opt.StopFollow:
			return nil
		default:
			buf := make([]byte, 1024)
			_, err := readCloser.Read(buf)
			if err != nil {
				return err
			}
			_, err = out.Write(buf)
			if err != nil {
				return err
			}
		}
	}
}
