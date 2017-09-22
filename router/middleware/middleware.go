package middleware

func SetNamespace(c *gin.Context) {
	ns := c.Param("namespace")
	if ns == "" {
		c.AbortWithErrorJSON(400, map[string]string{
			"error": `missing "namespace" parameter`,
		})
	}
}

func SetRandomKubeClient(c *gin.Context) {
	if len(server.KubeClients) == 0 {
		c.AbortWithErrorJSON(500, map[string]string{
			"error": "no available kubernetes apiserver clients"
		})
		return
	}

	n := rand.Intn(len(server.KubeClients))
	utils.Log(c).Infof("picked server.KubeClients[%d]", n)
	utils.AddLogField(c, "kubeclient-address", server.KubeClients[n].Tag)
	server.KubeClients[n].UseCount++
	c.Set("kubeclient", server.KubeClients[n])
}

// ParseJSON parses a JSON payload into a kubernetes struct of appropriate
// type and sets it into the gin context under the "requestObject" key.
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
		var obj *v1.Namespace
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	case "Deployment":
		var obj *v1beta2.Deployment
		err = json.Unmarshal(jsn, &obj)
		obj.Status = v1beta2.DeploymentStatus{}
		c.Set("requestObject", obj)
	case "Service":
		var obj *v1.Service
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	case "Endpoints":
		var obj *v1.Endpoints
		err = json.Unmarshal(jsn, &obj)
		c.Set("requestObject", obj)
	}
}
