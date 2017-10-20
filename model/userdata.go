package model

type UserData struct {
	ID     string `json:"id"`     // hosting-internal name
	Label  string `json:"label"`  // user-visible label for the object
	Access string `json:"access"` // one of: "owner", "read-delete"
}
