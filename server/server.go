package server

import (
	"k8s.io/client-go/kubernetes"
)

type Server struct {
	kubeclients []*kubernetes.Clientset
}

// GetKubeClient picks and returns a kubernetes.Clientset
// according to whatever load-balancing algorithm.
func (s *Server) GetKubeClient() *kubernetes.Clientset {
	// TODO
	return s.kubeclients[0]
}
