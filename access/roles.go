package access

type HTTPHeaders struct {
	Namespace []model.UserData
	Volume    []model.UserData
}

func (hhdr HTTPHeaders) CanCreateObject() bool {
}

func (hhdr HTTPHeaders) CanCreateNamespace() bool {
}

func (h HTTPHeaders) CanUseVolume(volLabel string) bool {
}
