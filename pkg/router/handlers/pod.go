package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

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

	rd, wr := io.Pipe()
	go kube.GetPodLogs(ns, ctx.Param(podParam), wr, &logOpt)
	go writeLogs(conn, rd)
}

func writeLogs(conn *websocket.Conn, logs *io.PipeReader) {
	defer conn.Close()
	bb := [logsBufferSize]byte{}
	buf := bytes.NewBuffer(bb[:])
	for {
		fmt.Println("Start reading logs")
		_, err := buf.ReadFrom(io.LimitReader(logs, logsBufferSize))
		if err != nil {
			if err == io.EOF || err == io.ErrClosedPipe {
				fmt.Println("Stop reading logs EOF/Closed pipe")
				return
			} else {
				log.WithError(err).Error("Unable read logs stream") //TODO: Write good err
				return
			}
			continue
		}
		fmt.Println("Start sending logs")
		if err := conn.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
			return
		}
		fmt.Println("Chunk sent")
		buf.Reset()
	}
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
