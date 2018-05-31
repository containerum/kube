package handlers

import (
	"net/http"
	"sync"
	"time"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	"git.containerum.net/ch/kube-api/pkg/utils/watchdog"
	"git.containerum.net/ch/kube-api/pkg/utils/wsutils"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	podParam       = "pod"
	containerQuery = "container"

	followQuery   = "follow"
	tailQuery     = "tail"
	previousQuery = "previous"

	ttyQuery         = "tty"
	interactiveQuery = "interactive"

	tailDefault = 100
	tailMax     = 1000

	wsBufferSize = 1024
	wsTimeout    = 5 * time.Second
	wsPingPeriod = time.Second
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  wsBufferSize,
	WriteBufferSize: wsBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// swagger:operation GET /namespaces/{namespace}/pods Pod GetPodList
// Get pods list.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
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
//  default:
//    $ref: '#/responses/error'
func GetPodList(ctx *gin.Context) {
	namespace := ctx.Param(namespaceParam)
	owner := ctx.Query(ownerQuery)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Owner":     owner,
	}).Debug("Get pod list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResourcesList()), ctx)
		return
	}

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
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
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
	namespace := ctx.Param(namespaceParam)
	podP := ctx.Param(podParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Pod":       podP,
	}).Debug("Get pod list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableGetResource()), ctx)
		return
	}

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
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
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
	namespace := ctx.Param(namespaceParam)
	podP := ctx.Param(podParam)
	log.WithFields(log.Fields{
		"Namespace": namespace,
		"Pod":       podP,
	}).Debug("Delete pod Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	_, err := kube.GetNamespace(namespace)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}

	err = kube.DeletePod(namespace, podP)
	if err != nil {
		gonic.Gonic(model.ParseKubernetesResourceError(err, kubeErrors.ErrUnableDeleteResource()), ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}

// swagger:operation GET /namespaces/{namespace}/pods/{pod}/log Pod GetPodLogs
// Get pod logs.
//
// ---
// x-method-visibility: public
// parameters:
//  - $ref: '#/parameters/UserIDHeader'
//  - $ref: '#/parameters/UserRoleHeader'
//  - $ref: '#/parameters/UserNamespaceHeader'
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
		"Namespace": ctx.Param(namespaceParam),
		"Pod":       ctx.Param(podParam),
		"Follow":    ctx.Query(followQuery),
		"Tail":      ctx.Query(tailQuery),
		"Container": ctx.Query(containerQuery),
		"Previous":  ctx.Query(previousQuery),
	}).Debug("Get pod logs Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	logOpt := makeLogOption(ctx)
	ns := ctx.Param(namespaceParam)

	rc, err := kube.GetPodLogs(ns, ctx.Param(podParam), &logOpt)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetPodLogs().AddDetailsErr(err), ctx)
		return
	}

	conn, err := wsupgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetPodLogs().AddDetailsErr(err), ctx)
		return
	}

	logStreamSetup(conn, rc, &logOpt)
}

func Exec(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace":   ctx.Param(namespaceParam),
		"Pod":         ctx.Param(podParam),
		"Container":   ctx.Query(containerQuery),
		"Interactive": ctx.Query(interactiveQuery),
		"TTY":         ctx.Query(ttyQuery),
	}).Debug("Exec Call")

	conn, err := wsupgrader.Upgrade(ctx.Writer, ctx.Request, ctx.Request.Header)
	if err != nil {
		ctx.Error(err)
		gonic.Gonic(kubeErrors.ErrUnableGetPodLogs().AddDetailsErr(err), ctx)
		return
	}

	conn.SetPongHandler(func(appData string) error {
		conn.SetWriteDeadline(time.Now().Add(wsTimeout))
		conn.SetReadDeadline(time.Now().Add(wsTimeout))
		return nil
	})
	conn.SetWriteDeadline(time.Now().Add(wsTimeout))
	conn.SetReadDeadline(time.Now().Add(wsTimeout))

	cmdMessage, err := receiveExecCommand(conn)
	if err != nil {
		wsutils.CloseWithCherry(conn, kubeErrors.ErrRequestValidationFailed().AddDetailsErr(err))
		return
	}

	opts, pipes, tsQueue := makeExecOptions(ctx, cmdMessage)

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var closeOnce sync.Once
	closeAll := func() {
		closeOnce.Do(func() {
			if pipes.StdinPipe != nil {
				pipes.StdinPipe.Close()
			}
			if pipes.StderrPipe != nil {
				pipes.StderrPipe.Close()
			}
			if pipes.StdoutPipe != nil {
				pipes.StdoutPipe.Close()
			}
			conn.WriteMessage(websocket.CloseMessage, nil)
		})
	}
	closeWd := watchdog.New(wsTimeout, func() { closeAll() })
	conn.SetPongHandler(func(appData string) error {
		conn.SetWriteDeadline(time.Now().Add(wsTimeout))
		conn.SetReadDeadline(time.Now().Add(wsTimeout))
		closeWd.Kick()
		return nil
	})

	err = kube.Exec(ctx.Param(namespaceParam), ctx.Param(podParam), opts)
	if err != nil {
		wsutils.CloseWithCherry(conn, kubeErrors.ErrUnableGetPodLogs().AddDetailsErr(err))
		return
	}

	go execFromClient(conn, tsQueue, pipes, closeAll)
	go execToClient(conn, pipes, closeAll)
}
