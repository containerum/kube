package handlers

import (
	"io"
	"strconv"
	"time"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/utils/timeoutreader"
	"git.containerum.net/ch/kube-api/pkg/utils/watchdog"
	"git.containerum.net/ch/kube-api/pkg/utils/wsutils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func logStreamSetup(conn *websocket.Conn, rc io.ReadCloser, logOpt *kubernetes.LogOptions) {
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
