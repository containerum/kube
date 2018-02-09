package model

type Protocol string

const (
	UDP Protocol = "UDP"
	TCP Protocol = "TCP"
)

type Service struct {
	Name      string        `json:"name" binding:"required"`
	Owner     string        `json:"owner" binding:"required"`
	CreatedAt int64         `json:"created_at, omitempty"`
	Deploy    string        `json:"deploy" binding:"required"`
	IP        []string      `json:"ip" omitempty"`
	Domain    string        `json:"domain, omitempty"`
	Type      string        `json:"type" binding:"required"`
	Ports     []ServicePort `json:"ports, omitempty"`
}

type ServicePort struct {
	Name       string   `json:"name" binding:"required"`
	Port       int      `json:"port" binding:"required"`
	TargetPort int      `json:"target_port, omitempty"`
	Protocol   Protocol `json:"protocol" binding:"required"`
}
