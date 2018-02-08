package model

type Protocol string

const (
	UDP Protocol = "UDP"
	TCP Protocol = "TCP"
)

// ServicePort is an user friendly service port representation
// Name is DNS_LABEL
// TargetPort is an int32 or IANA_SVC_NAME
// Protocol is TCP or UDP
type ServicePort struct {
	Name       string   `json:"name"`
	Port       uint32   `json:"port"`
	TargetPort uint32   `json:"target_port"`
	Protocol   Protocol `json:"protocol"`
}

// Service is an user friendly kebernetes service representation
// CreatedAt is an unix timestamp
type Service struct {
	CreatedAt int64         `json:"created_at"`
	Deploy    string        `json:"deploy"`
	IP        []string      `json:"ip"`
	Domain    string        `json:"domain, omitempty"`
	Type      string        `json:"type"`
	Ports     []ServicePort `json:"ports"`
}
