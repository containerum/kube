package handlers

import (
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

	"fmt"

	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	podParam       = "pod"
	followQuery    = "follow"
	tailQuery      = "tail"
	containerQuery = "container"
	previousQuery  = "previous"

	logsBufferSize = 1024
	tailDefault    = 100
	tailMax        = 1000
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  logsBufferSize,
	WriteBufferSize: logsBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func GetPodList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Owner":           ctx.Query(ownerQuery),
	}).Debug("Get pod list Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	pods, err := kube.GetPodList(ctx.MustGet(m.NamespaceKey).(string), ctx.Query(ownerQuery))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	podList := model.ParsePodList(pods, role == "user")
	ctx.JSON(http.StatusOK, podList)
}

func GetPod(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Pod":             ctx.Param(podParam),
	}).Debug("Get pod list Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	pod, err := kube.GetPod(ctx.MustGet(m.NamespaceKey).(string), ctx.Param(podParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	po := model.ParsePod(pod, role == "user")
	ctx.JSON(http.StatusOK, po)
}

func DeletePod(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Pod":       ctx.Param(podParam),
	}).Debug("Delete pod Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeletePod(ctx.Param(namespaceParam), ctx.Param(podParam))
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(model.ParseResourceError(err, cherry.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}

func GetPodLogs(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Pod":             ctx.Param(podParam),
		"Follow":          ctx.Query(followQuery),
		"Tail":            ctx.Query(tailQuery),
		"Container":       ctx.Query(containerQuery),
		"Previous":        ctx.Query(previousQuery),
	}).Debug("Get pod logs Call")

	conn, err := wsupgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.Error(err)
		log.WithError(err).Error("unable to upgrade http to socket")
		return
	}
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	logOpt := makeLogOption(ctx)
	ns := ctx.MustGet(m.NamespaceKey).(string)

	rc, err := kube.GetPodLogs(ns, ctx.Param(podParam), &logOpt)
	if err != nil {
		ctx.Error(err)
		return
	}
	go writeLogs(conn, rc)
}

func runOnce(action func()) func() {
	once := &sync.Once{}
	return func() {
		once.Do(action)
	}
}

func writeLogs(conn *websocket.Conn, logs io.ReadCloser) {
	defer conn.Close()
	pp := [logsBufferSize]byte{}
	const timeout = 4 * time.Second
	stop := make(chan struct{})
	stopAll := runOnce(func() {
		close(stop)
	})
	closeLogs := runOnce(func() {
		logs.Close()
	})
	defer closeLogs()
	defer stopAll()

	timer := time.NewTimer(timeout)

	conn.SetPongHandler(func(data string) error {
		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(timeout)
		err := conn.WriteControl(websocket.PingMessage,
			[]byte{},
			time.Now().Add(timeout))
		if err != nil {
			stopAll()
			closeLogs()
			log.Debugf("error while sending ping: %v", err)
			return err
		}
		return nil
	})
	err := conn.WriteControl(websocket.PingMessage,
		[]byte{},
		time.Now().Add(timeout))
	if err != nil {
		stopAll()
		closeLogs()
		log.Debugf("error while sending ping: %v", err)
		return
	}
	go func() {
		select {
		case <-stop:
			return
		case <-timer.C:
			log.Debugf("closing on timeout")
			stopAll()
			closeLogs()
		}
	}()

cycle:
	for {
		select {
		case <-stop:
			break cycle
		default:
			fmt.Println("Start reading logs")
			n, err := logs.Read(pp[:])
			log.Debugf("Read bytes from logs %v/n", n)
			switch err {
			case nil:
				// pass
			case io.EOF:
				break cycle
			default:
				log.WithError(err).Errorf("fatal error while reading logs from kube")
				break cycle
			}
			log.Debugf("Start sending logs")
			conn.SetWriteDeadline(time.Now().Add(3 * timeout / 4))
			err = conn.WriteMessage(websocket.TextMessage, pp[:n])
			switch err {
			case nil:
				// pass
			case websocket.ErrCloseSent:
				break cycle
			default:
				log.WithError(err).Debugf("error while sending log chunk to ws")
				break cycle
			}
			log.Debugf("Chunk sent")
		}
	}
	log.Debugf("End writing logs")
}

func makeLogOption(c *gin.Context) kubernetes.LogOptions {
	followStr := c.Query(followQuery)
	previousStr := c.Query(previousQuery)
	container := c.Query(containerQuery)
	tail, _ := strconv.Atoi(c.Query(tailQuery))
	if tail <= 0 || tail > tailMax {
		tail = tailDefault
	}
	return kubernetes.LogOptions{
		Tail:      int64(tail),
		Follow:    followStr == "true",
		Previous:  previousStr == "true",
		Container: container,
	}
}
