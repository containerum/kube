package model

type Namespace struct {
	Created   int64     `json:"created_at,omitempty"`
	Name      string    `json:"name" binding:"required"`
	Owner     string    `json:"owner,omitempty"`
	Resources Resources `json:"resources" binding:"required"`
}

type Resources struct {
	Hard Resource  `json:"hard" binding:"required"`
	Used *Resource `json:"used,omitempty"`
}

type Resource struct {
	CPU    string `json:"cpu" binding:"required"`
	Memory string `json:"memory" binding:"required"`
}

type UpdateNamespace struct {
	Resources Resources `json:"resources" binding:"required"`
}
