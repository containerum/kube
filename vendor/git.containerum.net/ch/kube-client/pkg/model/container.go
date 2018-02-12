package model

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
