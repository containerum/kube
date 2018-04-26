package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"git.containerum.net/ch/kube-api/pkg/utils/timeoutreader"
	"git.containerum.net/ch/kube-api/pkg/utils/watchdog"
	"git.containerum.net/ch/kube-api/pkg/utils/wsutils"

	"git.containerum.net/ch/cherry/adaptors/gonic"
	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
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

	wsTimeout    = 5 * time.Second
	wsPingPeriod = time.Second
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  logsBufferSize,
	WriteBufferSize: logsBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// swagger:operation GET /namespaces/{namespace}/pods Pod GetPodList
// Get pods list.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#get-pod-list
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: owner
//    in: query
//    type: string
//    required: false
// responses:
//  '200':
//    description: pod list
//    schema:
//      $ref: '#/definitions/PodsList'
//  configmap:
//    $ref: '#/responses/error'
func GetPodList(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	owner := ctx.Query(ownerQuery)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Owner":           owner,
	}).Debug("Get pod list Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	pods, err := kube.GetPodList(namespace, owner)
	if err != nil {
		gonic.Gonic(kubeErrors.ErrUnableGetResourcesList(), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	podList := model.ParseKubePodList(pods, role == m.RoleUser)
	ctx.JSON(http.StatusOK, podList)
}

// swagger:operation GET /namespaces/{namespace}/pods/{pod} Pod GetPod
// Get pod.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#get-pod
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: pod
//    in: query
//    type: string
//    required: true
// responses:
//  '200':
//    description: pod
//    schema:
//      $ref: '#/definitions/PodWithOwner'
//  default:
//    $ref: '#/responses/error'
func GetPod(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	podP := ctx.Param(podParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Pod":             podP,
	}).Debug("Get pod list Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	pod, err := kube.GetPod(namespace, podP)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

	role := ctx.MustGet(m.UserRole).(string)
	po := model.ParseKubePod(pod, role == m.RoleUser)
	ctx.JSON(http.StatusOK, po)
}

// swagger:operation DELETE /namespaces/{namespace}/pods/{pod} Pod DeletePod
// Delete pod.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#delete-pod
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - name: namespace
//    in: path
//    type: string
//    required: true
//  - name: pod
//    in: path
//    type: string
//    required: true
// responses:
//  '202':
//    description: pod deleted
//  default:
//    $ref: '#/responses/error'
func DeletePod(ctx *gin.Context) {
	namespace := ctx.MustGet(m.NamespaceKey).(string)
	podP := ctx.Param(podParam)
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       namespace,
		"Pod":             podP,
	}).Debug("Delete pod Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeletePod(namespace, podP)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}

// swagger:operation GET /namespaces/{namespace}/pods/{pod}/log Pod GetPodLogs
// Get pod logs.
// https://ch.pages.containerum.net/api-docs/modules/kube-api/index.html#delete-pod
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
//  - $ref: '#/parameters/UserVolumeHeader'
//  - $ref: '#/parameters/UpgradeHeader'
//  - $ref: '#/parameters/ConnectionHeader'
//  - $ref: '#/parameters/SecWebSocketKeyHeader'
//  - $ref: '#/parameters/SecWebsocketVersionHeader'
//  - name: namespace
//    in: header
//    type: string
//    required: true
//  - name: pod
//    in: path
//    type: string
//    required: true
//  - name: follow
//    in: query
//    type: bool
//    required: false
//  - name: tail
//    in: query
//    type: integer
//    required: false
//  - name: container
//    in: query
//    type: string
//    required: false
//  - name: previous
//    in: query
//    type: bool
//    required: false
// responses:
//  '101':
//    description: pod logs
//  default:
//    $ref: '#/responses/error'
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

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	logOpt := makeLogOption(ctx)
	ns := ctx.MustGet(m.NamespaceKey).(string)

	rc, err := kube.GetPodLogs(ns, ctx.Param(podParam), &logOpt)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetPodLogs().AddDetailsErr(err), ctx)
		return
	}

	conn, err := wsupgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(cherry.ErrUnableGetPodLogs().AddDetailsErr(err), ctx)
		return
	}

	if logOpt.Follow {
		rc = timeoutreader.NewTimeoutReader(rc, 30*time.Minute, true)
	} else {
		rc = timeoutreader.NewTimeoutReader(rc, 10*time.Second, true)
	}

	// watchdog for reader, resets by websocket pong
	closeWd := watchdog.New(wsTimeout, func() { rc.Close() })

	conn.SetPongHandler(func(appData string) error {
		conn.SetWriteDeadline(time.Now().Add(wsTimeout))
		conn.SetReadDeadline(time.Now().Add(wsTimeout))
		closeWd.Kick()
		return nil
	})
	conn.SetWriteDeadline(time.Now().Add(wsTimeout))
	conn.SetReadDeadline(time.Now().Add(wsTimeout))

	var (
		done = make(chan struct{})
		data = make(chan []byte)
	)
	go readConn(conn)
	go readLogs(rc, data, done)
	go writeLogs(conn, data, done)
}

func readLogs(logs io.ReadCloser, ch chan<- []byte, done chan<- struct{}) {
	buf := [logsBufferSize]byte{}
	defer func() { logs.Close(); done <- struct{}{} }()

	for {
		readBytes, err := logs.Read(buf[:])
		switch err {
		case nil:
			// pass
		case io.EOF, timeoutreader.ErrReadTimeout:
			return
		default:
			log.WithError(err).Error("Log read failed")
			return
		}

		ch <- buf[:readBytes]
	}
}

func writeLogs(conn *websocket.Conn, ch <-chan []byte, done <-chan struct{}) {
	defer conn.Close()
	pingTimer := time.NewTicker(wsPingPeriod)

	for {
		var err error
		select {
		case <-done:
			return
		case <-pingTimer.C:
			err = conn.WriteMessage(websocket.PingMessage, nil)
		case data := <-ch:
			err = conn.WriteMessage(websocket.TextMessage, data)
		}

		switch {
		case err == nil,
			wsutils.IsNetTemporary(err):
			// pass
		case err == timeoutreader.ErrReadTimeout,
			err == websocket.ErrCloseSent, // connection closed by us
			wsutils.IsNetTimeout(err),     // deadline failed
			wsutils.IsBrokenPipe(err),     // connection closed by client
			wsutils.IsClose(err):
			return
		default:
			log.WithError(err).Errorf("Log send failed")
			return
		}
	}
}

func readConn(conn *websocket.Conn) {
	for {
		_, _, err := conn.ReadMessage() // to trigger pong handlers and check connection though
		if err != nil {
			conn.Close()
			return
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
