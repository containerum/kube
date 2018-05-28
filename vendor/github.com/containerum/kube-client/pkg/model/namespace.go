package model

// Resources -- represents namespace limits and user resources.
//
// swagger:model
type Resources struct {
	// Hard resource limits
	//
	// required: true
	Hard Resource  `json:"hard"`
	Used *Resource `json:"used,omitempty"`
}

// Resource -- represents namespace CPU and RAM
//
// swagger:model
type Resource struct {
	// CPU in m
	//
	// required: true
	CPU uint `json:"cpu"`
	// RAM in Mi
	//
	// required: true
	Memory uint `json:"memory"`
}

// UpdateNamespaceName -- contains new namespace label
//
// swagger:model
type UpdateNamespaceName struct {
	// required: true
	Label string `json:"label"`
}

// Namespace -- namespace representation
//
// swagger:model
type Namespace struct {
	//creation date in RFC3339 format
	CreatedAt *string `json:"created_at,omitempty"`
	ID        string  `json:"id,omitempty"`
	// user-visible label for the namespace
	Label         string   `json:"label,omitempty"`
	Access        string   `json:"access,omitempty"`
	MaxExtService *uint    `json:"max_ext_service,omitempty"`
	MaxIntService *uint    `json:"max_int_service,omitempty"`
	MaxTraffic    *uint    `json:"max_traffic,omitempty"`
	Volumes       []Volume `json:"volumes,omitempty"`
	// required: true
	Resources Resources `json:"resources"`
}
