package kubernetes

import (
	"io"
	"time"

	"git.containerum.net/ch/kube-api/pkg/utils"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
)

const (
	logsBufferSize = 1024
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
		}).Error(ErrUnableGetPodList)
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
		}).Error(ErrUnableGetPod)
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
		}).Error(ErrUnableDeletePod)
		return err
	}
	return nil
}

//GetPodLogs attaches client to pod log
func (k *Kube) GetPodLogs(ns string, po string, opt *LogOptions) (io.ReadCloser, error) {
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
		return nil, err
	}
	return timeoutreader.NewTimeoutReaderSize(readCloser, 10*time.Second, true, 0), nil
}

type proxyRC struct {
	io.ReadCloser
}

func (proxy *proxyRC) Read(p []byte) (int, error) {
	n, err := proxy.ReadCloser.Read(p)
	if n == 0 && err != nil {
		return 0, io.EOF
	}
	return n, err
}

func (proxy *proxyRC) Close() error {
	return proxy.ReadCloser.Close()
}
