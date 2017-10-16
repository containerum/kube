package access

import (
	"bitbucket.org/exonch/kube-api/model"
	//"k8s.io/api/apps/v1beta1"
	//"k8s.io/api/core/v1"
)

type HTTPHeaders struct {
	Namespace []model.UserData
	Volume    []model.UserData
}

/*
	// compares UserData.Label and x.metadata.name
	func (h HTTPHeaders) CanCreateObject(x interface{}) bool {
		var ans bool
		var xlabel string

		switch xt := x.(type) {
		case *v1.Namespace:
			xlabel = xt.ObjectMeta.Name
		case *v1.Service:
			xlabel = xt.ObjectMeta.Name
		case *v1.Secret:
			xlabel = xt.ObjectMeta.Name
		case *v1beta1.Deployment:
			xlabel = xt.ObjectMeta.Name
		}
		for i, ns := range h.Namespace {
			if ns.Label == xlabel {
				switch ns.Access {
				case "owner":
					ans = true
				case "read-delete":
					fallthrough
				default:
					ans = false
				}
				break
			}
		}
		return ans
	}

	func (h HTTPHeaders) CanUseVolume(volLabel string) bool {
		//
	}

	func (h HTTPHeaders) CanDeleteObject(x interface{}) bool {
		var ans bool
		var xlabel string

		switch xt := x.(type) {
		case *v1.Namespace:
			xlabel = xt.ObjectMeta.Name
		case *v1.Service:
			xlabel = xt.ObjectMeta.Name
		case *v1.Secret:
			xlabel = xt.ObjectMeta.Name
		case *v1beta1.Deployment:
			xlabel = xt.ObjectMeta.Name
		}
		for i, ns := range h.Namespace {
			if ns.Label == xlabel {
				switch ns.Access {
				case "owner", "read-delete":
					ans = true
				default:
					ans = false
				}
				break
			}
		}
		return ans
	}
*/
