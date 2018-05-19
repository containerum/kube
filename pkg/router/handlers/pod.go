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
	"git.containerum.net/ch/kube-api/pkg/utils/terminal"
	"git.containerum.net/ch/kube-api/pkg/utils/timeoutreader"
	"git.containerum.net/ch/kube-api/pkg/utils/watchdog"
	"git.containerum.net/ch/kube-api/pkg/utils/wsutils"
	"git.containerum.net/ch/kube-api/proto"
	"github.com/gogo/protobuf/proto"
	"k8s.io/client-go/tools/remotecommand"

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
	buf := [wsBufferSize]byte{}
	defer logs.Close()
	defer func() { done <- struct{}{} }()

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
	defer func() {
		conn.WriteMessage(websocket.CloseMessage, nil)
		conn.Close()
	}()
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
	defer log.Debugf("End writing logs")
	for {
		_, _, err := conn.ReadMessage() // to trigger pong handlers and check connection though
		if err != nil {
			conn.Close()
			return
		}
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

func receiveExecCommand(conn *websocket.Conn) (cmdMessage kubeProto.ExecCommand, err error) {
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.BinaryMessage {
			continue
		}
		err = proto.Unmarshal(message, &cmdMessage)
		return
	}
}

type execPipes struct {
	StdoutPipe, StderrPipe io.ReadCloser
	StdinPipe              io.WriteCloser
}

func makeExecOptions(ctx *gin.Context, cmdMessage kubeProto.ExecCommand) (opts *kubernetes.ExecOptions, pipes *execPipes, tsQueue *terminal.SizeQueue) {
	command := append([]string{cmdMessage.GetCommand()}, cmdMessage.GetArgs()...)

	tty := ctx.Query(ttyQuery) == "true"

	interactive := ctx.Query(interactiveQuery) == "true"

	var (
		stderrIn, stderrOut = io.Pipe()
		stdoutIn, stdoutOut = io.Pipe()
		stdinIn             *io.PipeReader
		stdinOut            *io.PipeWriter
	)

	if interactive {
		stdinIn, stdinOut = io.Pipe()
	}

	pipes = &execPipes{
		StdoutPipe: stdoutIn,
		StderrPipe: stderrIn,
		StdinPipe:  stdinOut,
	}

	tsQueue = terminal.NewSizeQueue(20)

	opts = &kubernetes.ExecOptions{
		Container:         ctx.Query(containerQuery),
		Command:           command,
		TTY:               tty,
		TerminalSizeQueue: tsQueue,
		Stderr:            stderrOut,
		Stdout:            stdoutOut,
		Stdin:             stdinIn,
	}

	return
}

func execFromClient(conn *websocket.Conn, tsQueue *terminal.SizeQueue, pipes *execPipes, closeAll func()) {
	defer closeAll()

	for {
		messageType, message, err := conn.ReadMessage()
		switch {
		case err == nil,
			wsutils.IsNetTemporary(err):
			// pass
		case wsutils.IsNetTimeout(err),
			wsutils.IsBrokenPipe(err),
			wsutils.IsClose(err):
			return
		default:
			log.WithError(err).Errorf("exec data read failed")
			return
		}

		if messageType != websocket.BinaryMessage {
			continue
		}

		var execFromClientMsg kubeProto.ExecFromClient
		err = proto.Unmarshal(message, &execFromClientMsg)
		if err != nil {
			log.WithError(err).Warnf("invalid exec data from client")
			continue
		}

		switch execFromClientMsg.ClientMessage.(type) {
		case *kubeProto.ExecFromClient_TerminalSize:
			if tsize := execFromClientMsg.GetTerminalSize(); tsize != nil {
				tsQueue.Put(remotecommand.TerminalSize{
					Width:  uint16(tsize.Width),
					Height: uint16(tsize.Height),
				})
			}
		case *kubeProto.ExecFromClient_StdinData:
			if pipes.StdinPipe != nil {
				_, err = pipes.StdinPipe.Write(execFromClientMsg.GetStdinData())
				if err != nil {
					return
				}
			}
		}
	}
}

func readToChan(rd io.Reader, data chan<- []byte, done chan<- struct{}, ack <-chan struct{}) {
	var buf [wsBufferSize]byte
	for {
		n, err := rd.Read(buf[:])
		if err != nil {
			done <- struct{}{}
			return
		}
		data <- buf[:n]
		<-ack
	}
}

func execToClient(conn *websocket.Conn, pipes *execPipes, closeAll func()) {
	defer closeAll()

	done := make(chan struct{})

	stderrData := make(chan []byte)
	stderrAck := make(chan struct{})

	stdoutData := make(chan []byte)
	stdoutAck := make(chan struct{})

	go readToChan(pipes.StderrPipe, stderrData, done, stderrAck)
	go readToChan(pipes.StdoutPipe, stdoutData, done, stdoutAck)

	pingTimer := time.NewTicker(wsPingPeriod)

	for {
		var err error
		select {
		case stderr := <-stderrData:
			rawMsg, marshalErr := proto.Marshal(&kubeProto.ExecToClient{
				ServerMessage: &kubeProto.ExecToClient_StderrData{StderrData: stderr},
			})
			if marshalErr != nil {
				return
			}
			err = conn.WriteMessage(websocket.BinaryMessage, rawMsg)
			stderrAck <- struct{}{}
		case stdout := <-stdoutData:
			rawMsg, marshalErr := proto.Marshal(&kubeProto.ExecToClient{
				ServerMessage: &kubeProto.ExecToClient_StdoutData{StdoutData: stdout},
			})
			if marshalErr != nil {
				return
			}
			err = conn.WriteMessage(websocket.BinaryMessage, rawMsg)
			stdoutAck <- struct{}{}
		case <-pingTimer.C:
			err = conn.WriteMessage(websocket.PingMessage, nil)
		case <-done:
			return
		}

		switch {
		case err == nil,
			wsutils.IsNetTemporary(err):
			// pass
		case err == timeoutreader.ErrReadTimeout,
			err == websocket.ErrCloseSent,
			wsutils.IsNetTimeout(err),
			wsutils.IsBrokenPipe(err),
			wsutils.IsClose(err):
			return
		default:
			log.WithError(err).Errorf("exec data send failed")
			return
		}
	}
}
