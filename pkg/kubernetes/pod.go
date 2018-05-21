package kubernetes

import (
	"io"
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
)

//TODO: Imp struct to GetPodLogs func
type LogOptions struct {
	Follow    bool
	Tail      int64
	Previous  bool
	Container string
}

type ExecOptions struct {
	Container         string
	Command           []string
	Stdin             io.Reader
	Stdout, Stderr    io.Writer
	TTY               bool
	TerminalSizeQueue remotecommand.TerminalSizeQueue
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

func (k *Kube) Exec(ns string, po string, opt *ExecOptions) error {
	// logic taken from "kubectl exec" command
	pod, err := k.CoreV1().Pods(ns).Get(po, meta_v1.GetOptions{})
	if err != nil {
		return err
	}

	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return kubeErrors.ErrRequestValidationFailed().
			AddDetailF("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
	}

	containerName := opt.Container
	if len(containerName) == 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	req := k.RESTClient().
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", containerName)
	req.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   opt.Command,
		Stdin:     opt.Stdin != nil,
		Stdout:    opt.Stdout != nil,
		Stderr:    opt.Stderr != nil,
		TTY:       opt.TTY,
	}, legacyscheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(k.config, http.MethodPost, req.URL())
	if err != nil {
		return err
	}
	return executor.Stream(remotecommand.StreamOptions{
		Stdin:             opt.Stdin,
		Stdout:            opt.Stdout,
		Stderr:            opt.Stderr,
		Tty:               opt.TTY,
		TerminalSizeQueue: opt.TerminalSizeQueue,
	})
}
