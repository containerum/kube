package router

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	podParam    = "pod"
	followQuery = "follow"
	tailQuery   = "tail"

	logsBufferSize = 1024
	tailDefault    = 200
	tailMax        = 1000
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func getPodList(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
		"Owner":     c.Query(ownerQuery),
	}).Debug("Get pod list Call")
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	pods, err := kube.GetPodList(c.Param(namespaceParam), c.Query(ownerQuery))
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	podList := model.ParsePodList(pods)
	c.JSON(http.StatusOK, podList)
}

func getPodLogs(c *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": c.Param(namespaceParam),
		"Pod":       c.Param(podParam),
	}).Debug("Get pod logs Call")

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.WithError(err).Error("unable to upgrade http to socket")
		return
	}

	stream := new(bytes.Buffer)
	kube := c.MustGet(m.KubeClient).(*kubernetes.Kube)
	stop := make(chan struct{}, 1)
	tail, _ := strconv.Atoi(c.Query(tailQuery))
	if tail <= 0 || tail > tailMax {
		tail = tailDefault
	}
	go kube.GetPodLogs(c.Param(namespaceParam), c.Param(podParam), stream, &kubernetes.LogOptions{
		Follow: c.Query(followQuery) == "true",
	})
	go writeLogs(conn, stream, &stop)
}

func writeLogs(conn *websocket.Conn, logs *bytes.Buffer, done *chan struct{}) {
	defer func(done *chan struct{}) {
		conn.Close()
		*done <- struct{}{}
	}(done)

	for {
		if logs == nil {
			continue
		}
		buf := make([]byte, logsBufferSize)
		_, err := logs.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.WithError(err).Error("Unable read logs stream") //TODO: Write good err
			}
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, buf); err != nil {
			return
		}
	}
}
