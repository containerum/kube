package terminal

import "k8s.io/client-go/tools/remotecommand"

type SizeQueue struct {
	sizes chan remotecommand.TerminalSize
}

func NewSizeQueue(queueLength int) *SizeQueue {
	return &SizeQueue{
		sizes: make(chan remotecommand.TerminalSize, queueLength),
	}
}

func (s *SizeQueue) Next() *remotecommand.TerminalSize {
	sz := <-s.sizes
	return &sz
}

func (s *SizeQueue) Put(size remotecommand.TerminalSize) {
	s.sizes <- size
}
