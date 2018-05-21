package handlers

import (
	"io"
	"time"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/utils/terminal"
	"git.containerum.net/ch/kube-api/pkg/utils/timeoutreader"
	"git.containerum.net/ch/kube-api/pkg/utils/wsutils"
	"git.containerum.net/ch/kube-api/proto"
	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/remotecommand"
)

type execPipes struct {
	StdoutPipe, StderrPipe io.ReadCloser
	StdinPipe              io.WriteCloser
}

func receiveExecCommand(conn *websocket.Conn) (cmdMessage kubeProto.ExecCommand, err error) {
	for {
		var (
			messageType int
			message     []byte
		)
		messageType, message, err = conn.ReadMessage()
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
