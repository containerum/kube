package main

import (
	"github.com/gin-gonic/gin"
)

const (
	envPrefix = "CH_KUBE_API_"
)

func getenv(n string) string {
	return os.Getenv(envPrefix + n)
}

func makeKubeClients() []*kubernetes.Clientset {
	res := []*kubernetes.Clientset{}
	for i := 0; ; i++ {
		kubeConfigPath := getenv(fmt.Sprintf("KUBECONFIG_%d", i))
		if kubeConfigPath == "" {
			break
		}

		config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			logrus.Fatalf("failed to parse kube config: %v", err)
		}
		kcli, err := kubernetes.NewForConfig(config)
		if err != nil {
			logrus.Fatalf("failed to create kube client from %s: %v",
				kubeConfigPath, err)
		}
		res = append(res, kcli)
	}
	return res
}

func main() {
	srv := &server.Server{
		KubeClients: makeKubeClients(),
	}
	g := gin.New()
	g.Use(http.InitCmdContext)
	g.POST("/namespaces", http.GinHandler(http.NamespaceCreate))
}
