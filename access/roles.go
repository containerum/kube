package access

import (
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
)

type HTTPHeaders struct {
	Namespace []model.UserData
	Volume    []model.UserData
}

func (h HTTPHeaders) CanCreateObject(x interface{}) bool {
	switch xt := x.(type) {
	case *v1.Namespace:
	case *v1.Service:
	case *v1.Secret:
	case *v1beta1.Deployment:
	}
}

func (h HTTPHeaders) CanUseVolume(volLabel string) bool {
}

func (h HTTPHeaders) CanDeleteObject(x interface{}) bool {
	switch xt := x.(type) {
	case *v1.Namespace:
	case *v1.Service:
	case *v1.Secret:
	case *v1beta1.Deployment:
	}
}
