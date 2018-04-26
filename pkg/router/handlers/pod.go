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
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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
//    description: error
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
//    description: error
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
//    description: error
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
//    description: error
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
	stream := new(bytes.Buffer)
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	logOpt := makeLogOption(ctx)
	ns := ctx.MustGet(m.NamespaceKey).(string)

	go kube.GetPodLogs(ns, ctx.Param(podParam), stream, &logOpt)
	go writeLogs(conn, stream, &logOpt.StopFollow, logOpt.Follow)
}

func writeLogs(conn *websocket.Conn, logs *bytes.Buffer, done *chan struct{}, follow bool) {
	defer func(done *chan struct{}) {
		conn.Close()
		*done <- struct{}{}
	}(done)

	for {
		time.Sleep(time.Millisecond * 5)
		buf := make([]byte, logsBufferSize)
		_, err := logs.Read(buf)
		if err != nil {
			if err == io.EOF {
				if !follow { // if we are not following logs just close connection
					return
				}
			} else {
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
