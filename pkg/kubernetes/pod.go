package kubernetes

import (
	"bytes"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"io"

	"fmt"

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
func (k *Kube) GetPodLogs(ns string, po string, out *io.PipeWriter, opt *LogOptions) error {
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
	bb := [logsBufferSize]byte{}
	buf := bytes.NewBuffer(bb[:])
	for {
		fmt.Println("Start reading log from stream")
		_, err := buf.ReadFrom(io.LimitReader(readCloser, logsBufferSize))
		if err != nil {
			out.CloseWithError(err)
			return err
		}
		n, err := buf.WriteTo(out)
		switch err {
		case nil:
			if n == 0 {
				return nil
			}
			//pass
		case io.ErrClosedPipe:
			log.Debugln("Connection closed")
			return nil
		default:
			fmt.Println("Error", err)
			out.CloseWithError(err)
			return err
		}
		buf.Reset()
		fmt.Println("Chunk of logs read end")
	}
}
