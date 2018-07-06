package wsutils

import (
	"net"
	"os"
	"syscall"

	"github.com/gorilla/websocket"
)

func IsNetTimeout(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}

func IsBrokenPipe(err error) bool {
	opErr, isOpErr := err.(*net.OpError)
	if !isOpErr {
		return false
	}
	syscallErr, ok := opErr.Err.(*os.SyscallError)
	return ok && syscallErr.Err == syscall.EPIPE
}

func IsClose(err error) bool {
	_, isClose := err.(*websocket.CloseError)
	if isClose {
		return true
	}
	opErr, isOpErr := err.(*net.OpError)
	if !isOpErr {
		return false
	}
	return opErr.Err.Error() == "use of closed network connection"
}

func IsNetTemporary(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Temporary()
}
