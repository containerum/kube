package server

// import (
// 	"k8s.io/client-go/kubernetes"
// )
//
// type Server struct {
// 	KubeClients []*kubernetes.Clientset
// }
//
// // GetKubeClient picks and returns a kubernetes.Clientset
// // according to whatever load-balancing algorithm.
// func (s *Server) GetKubeClient() *kubernetes.Clientset {
// 	// TODO
// 	return s.kubeclients[0]
// }

// parseJSON parses a JSON payload into a kubernetes struct.
func ParseJSON(c *gin.Context) {
	var kind struct {
		Kind string `json:"kind"`
	}

	jsn, err := c.GetRawData()
	if err != nil {
		c.AbortWithStatusJSON(999, map[string]string{"error": "bad request"})
		return
	}

	err = json.Unmarshal(jsn, &kind)
	if err != nil {
		return nil, "", err
	}
	switch kind.Kind {
	case "Namespace":
		var obj *corev1.Namespace
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	case "Deployment":
		var obj *appsv1beta2.Deployment
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	case "Service":
		var obj *corev1.Service
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	case "Endpoints":
		var obj *corev1.Endpoints
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	}
}
