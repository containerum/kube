package model

type Deployment struct {
	Name            string              `json:"name" binding:"required"`
	Owner           string              `json:"owner_id" binding:"required"`
	Replicas        int                 `json:"replicas" binding:"required"`
	Containers      *[]Container        `json:"containers" binding:"required"`
	ImagePullSecret map[string]string   `json:"image_pull_secret,omitempty"`
	Status          *DeploymentStatus   `json:"status,omitempty"`
	Hostname        *string             `json:"hostname,omitempty"`
	Volume          *[]DeploymentVolume `json:"volume,omitempty"`
}

type DeploymentVolume struct {
	Name string `json:"name" binding:"required"`
	GlusterFS GlusterFS `json:"glusterfs" binding:"required"`
}

type GlusterFS struct {
	Endpoint string `json:"endpoint" binding:"required"`
	Path     string `json:"path" binding:"required"`
}


type DeploymentStatus struct {
	Created             int64 `json:"created_at"`
	Updated             int64 `json:"updated_at"`
	Replicas            int   `json:"replicas"`
	ReadyReplicas       int   `json:"ready_replicas"`
	AvailableReplicas   int   `json:"available_replicas"`
	UnavailableReplicas int   `json:"unavailable_replicas"`
	UpdatedReplicas     int   `json:"updated_replicas"`
}

type UpdateReplicas struct {
	Replicas int `json:"replicas" binding:"required"`
}
