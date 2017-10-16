package access

import (
	"github.com/gin-gonic/gin"
)

type Perm string

const (
	Create Perm = "create"
	Delete Perm = "delete"
	List   Perm = "list"
	Read   Perm = "read"
	Edit   Perm = "edit"
)

// accessMap : objtype → Perm → role → true or false
var accessMap = make(map[string]map[Perm]map[string]bool)

func init() {
	objtypes := []string{
		"Namespace",
		"Deployment",
		"Service",
		"Endpoints",
		"ConfigMap",
		"Secret",
	}
	perms := []Perm{
		Create,
		Delete,
		List,
		Read,
		Edit,
	}
	accessLevels := []string{
		"owner",
		"read",
		"write",
		"read-delete",
	}

	for _, x := range objtypes {
		accessMap[x] = make(map[Perm]map[string]bool)
		for _, y := range perms {
			accessMap[x][y] = make(map[string]bool)
			for _, z := range accessLevels {
				accessMap[x][y][z] = false
			}
		}
	}
}

// CheckAccess returns a gin.HandlerFunc that checks for specified
// access permissions in the supplied HTTP headers.
//
// objtype is a kubernetes object kind, e.g. "DeploymentList", "Namespace".
// perm is one of "read", "write", "delete".
func CheckAccess(objtype string, perm Perm) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
