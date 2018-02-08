package model

type Secret struct {
	Data map[string]string `json:"data" binding:"required"`
	Name string            `json:"name" binding:"required"`
}
