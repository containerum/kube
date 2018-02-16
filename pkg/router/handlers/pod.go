package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

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
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func GetPodList(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": c.Param(namespaceParam),
		"Namespace":       c.MustGet(m.NamespaceKey).(string),
		"Owner":           c.Query(ownerQuery),
	}).Debug("Get pod list Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	pods, err := kube.GetPodList(c.MustGet(m.NamespaceKey).(string), c.Query(ownerQuery))
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	podList := model.ParsePodList(pods)
	c.JSON(http.StatusOK, podList)
}

func GetPod(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": c.Param(namespaceParam),
		"Namespace":       c.MustGet(m.NamespaceKey).(string),
		"Pod":             c.Param(podParam),
	}).Debug("Get pod list Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	pod, err := kube.GetPod(c.MustGet(m.NamespaceKey).(string), c.Param(podParam))
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	po := model.ParsePod(pod)
	c.JSON(http.StatusOK, po)
}

func DeletePod(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
		"Pod":       c.Param(podParam),
	}).Debug("Delete pod Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeletePod(c.Param(namespaceParam), c.Param(podParam))
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	c.Status(http.StatusAccepted)
}

func GetPodLogs(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
		"Pod":       c.Param(podParam),
		"Follow":    c.Query(followQuery),
		"Tail":      c.Query(tailQuery),
		"Container": c.Query(containerQuery),
		"Previous":  c.Query(previousQuery),
	}).Debug("Get pod logs Call")

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.WithError(err).Error("unable to upgrade http to socket")
		return
	}
	stream := new(bytes.Buffer)
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	logOpt := makeLogOption(c)

	go kube.GetPodLogs(c.Param(namespaceParam), c.Param(podParam), stream, &logOpt)
	go writeLogs(conn, stream, &logOpt.StopFollow)
}

func writeLogs(conn *websocket.Conn, logs *bytes.Buffer, done *chan struct{}) {
	defer func(done *chan struct{}) {
		conn.Close()
		*done <- struct{}{}
	}(done)

	for {
		time.Sleep(time.Millisecond * 5)
		buf := make([]byte, logsBufferSize)
		_, err := logs.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.WithError(err).Error("Unable read logs stream") //TODO: Write good err
				return
			}
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, buf); err != nil {
			return
		}
	}
}

func makeLogOption(c *gin.Context) kubernetes.LogOptions {
	stop := make(chan struct{}, 1)
	followStr := c.Query(followQuery)
	previousStr := c.Query(previousQuery)
	container := c.Query(containerQuery)
	tail, _ := strconv.Atoi(c.Query(tailQuery))
	if tail <= 0 || tail > tailMax {
		tail = tailDefault
	}
	return kubernetes.LogOptions{
		Tail:       int64(tail),
		Follow:     followStr == "true",
		StopFollow: stop,
		Previous:   previousStr == "true",
		Container:  container,
	}
}
