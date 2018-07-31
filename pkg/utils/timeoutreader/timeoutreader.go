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
	timeout        time.Duration
	closeOnTimeout bool
	maxReadSize    int
	done           chan *readResponse
	timer          *time.Timer
	close          chan struct{}
	onceClose      sync.Once
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
func (this *TimeoutReader) Close() (err error) {
	this.onceClose.Do(func() {
		this.close <- struct{}{}
		err = this.reader.Close()
	})
	return
}

// Read from the underlying reader.
// If the underlying Read() does not return within the timeout, ErrReadTimeout
// is returned.
func (this *TimeoutReader) Read(p []byte) (int, error) {
	if this.timeout <= 0 {
		return this.reader.Read(p)
	}

	if this.maxReadSize > 0 && len(p) > this.maxReadSize {
		p = p[:this.maxReadSize]
	}

	// reset the timer
	select {
	case <-this.timer.C:
	default:
	}
	this.timer.Reset(this.timeout)

	// clear the done channel
	select {
	case <-this.done:
	default:
	}

	var timedOut bool
	var finished bool
	var mutex sync.Mutex

	go func() {
		n, err := io.ReadAtLeast(this.reader, p, 1)
		mutex.Lock()
		defer mutex.Unlock()
		finished = true
		if !timedOut {
			this.timer.Stop()
			if err == io.ErrUnexpectedEOF {
				err = nil
			}
			this.done <- &readResponse{n: n, err: err}
		}
	}()

	select {
	case <-this.timer.C:
		mutex.Lock()
		defer mutex.Unlock()
		if finished {
			resp := <-this.done
			return resp.n, resp.err
		}
		timedOut = true
		if this.closeOnTimeout {
			this.reader.Close()
		}
		return 0, ErrReadTimeout
	case <-this.close:
		return 0, io.EOF
	case resp := <-this.done:
		return resp.n, resp.err
	}
}
