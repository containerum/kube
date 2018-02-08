package model

type Pod struct {
	Name            string            `json:"name" binding:"required"`
	Owner           *string           `json:"owner_id,omitempty"`
	Containers      []Container       `json:"containers"`
	ImagePullSecret map[string]string `json:"image_pull_secret,omitempty"`
	Status          *PodStatus        `json:"status,omitempty"`
	Hostname        *string           `json:"hostname,omitempty"`
}

type PodStatus struct {
	Phase string `json:"phase"`
}

type Container struct {
	Name    string    `json:"name" binding:"required"`
	Env     *[]Env    `json:"env,omitempty"`
	Image   string    `json:"image" binding:"required"`
	Volume  *[]Volume `json:"volume,omitempty"`
	Limits  Limits    `json:"limits" binding:"required"`
	Ports   *[]Port   `json:"ports,omitempty"`
	Command *[]string `json:"command,omitempty"`
}

type Env struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type Volume struct {
	Name      string  `json:"name" binding:"required"`
	MountPath string  `json:"mount_path" binding:"required"`
	SubPath   *string `json:"sub_path,omitempty"`
}

type Limits struct {
	CPU    string `json:"cpu" binding:"required"`
	Memory string `json:"memory" binding:"required"`
}

type Port struct {
	Name     *string `json:"name,omitempty""`
	Port     int32   `json:"containerPort" binding:"required"`
	Protocol string  `json:"protocol" binding:"required"`
}

type UpdateImage struct {
	ContainerName string `json:"container" binding:"required"`
	Image         string `json:"image" binding:"required"`
}
