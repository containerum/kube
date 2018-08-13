package timeoutreader

import (
	"errors"
	"io"
	"sync"
	"time"
)

// ErrReadTimeout is the error used when a read times out before completing.
var ErrReadTimeout = errors.New("read timed out")

type readResponse struct {
	n   int
	err error
}

// An io.ReadCloser that has a timeout for each underlying Read() function and
// optionally closes the underlying Reader on timeout.
type TimeoutReader struct {
	reader         io.ReadCloser
	done           chan *readResponse
	timer          *time.Timer
	close          chan struct{}
	onceClose      sync.Once
	timeout        time.Duration
	maxReadSize    int
	closeOnTimeout bool
}

func NewTimeoutReaderSize(reader io.ReadCloser, timeout time.Duration, closeOnTimeout bool, maxReadSize int) *TimeoutReader {
	tr := new(TimeoutReader)
	tr.reader = reader
	tr.timeout = timeout
	tr.closeOnTimeout = closeOnTimeout
	tr.maxReadSize = maxReadSize
	tr.done = make(chan *readResponse, 1)
	if timeout > 0 {
		tr.timer = time.NewTimer(timeout)
	}
	tr.close = make(chan struct{})
	return tr
}

// Create a new TimeoutReader.
func NewTimeoutReader(reader io.ReadCloser, timeout time.Duration, closeOnTimeout bool) *TimeoutReader {
	return NewTimeoutReaderSize(reader, timeout, closeOnTimeout, 0)
}

// Closes the TimeoutReader.
// Also closes the underlying Reader if it was not closed already at timeout.
func (tr *TimeoutReader) Close() (err error) {
	tr.onceClose.Do(func() {
		tr.close <- struct{}{}
		err = tr.reader.Close()
	})
	return
}

// Read from the underlying reader.
// If the underlying Read() does not return within the timeout, ErrReadTimeout
// is returned.
func (tr *TimeoutReader) Read(p []byte) (int, error) {
	if tr.timeout <= 0 {
		return tr.reader.Read(p)
	}

	if tr.maxReadSize > 0 && len(p) > tr.maxReadSize {
		p = p[:tr.maxReadSize]
	}

	// reset the timer
	select {
	case <-tr.timer.C:
	default:
	}
	tr.timer.Reset(tr.timeout)

	// clear the done channel
	select {
	case <-tr.done:
	default:
	}

	var timedOut bool
	var finished bool
	var mutex sync.Mutex

	go func() {
		n, err := io.ReadAtLeast(tr.reader, p, 1)
		mutex.Lock()
		defer mutex.Unlock()
		finished = true
		if !timedOut {
			tr.timer.Stop()
			if err == io.ErrUnexpectedEOF {
				err = nil
			}
			tr.done <- &readResponse{n: n, err: err}
		}
	}()

	select {
	case <-tr.timer.C:
		mutex.Lock()
		defer mutex.Unlock()
		if finished {
			resp := <-tr.done
			return resp.n, resp.err
		}
		timedOut = true
		if tr.closeOnTimeout {
			tr.reader.Close()
		}
		return 0, ErrReadTimeout
	case <-tr.close:
		return 0, io.EOF
	case resp := <-tr.done:
		return resp.n, resp.err
	}
}
