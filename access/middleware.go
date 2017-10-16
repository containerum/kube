package access

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

// Required permission for the action on the object.
type Perm string

const (
	Create Perm = "create"
	Delete Perm = "delete"
	List   Perm = "list"
	Read   Perm = "read"
	Edit   Perm = "edit"
)

// User's access level for this object.
type AccessLevel string

const (
	LvlOwner      AccessLevel = "owner"
	LvlRead       AccessLevel = "read"
	LvlWrite      AccessLevel = "write"
	LvlReadDelete AccessLevel = "read-delete"
)

// accessMap : objtype → Perm → role → (true|false)
var accessMap = make(map[string]map[Perm]map[AccessLevel]bool)

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
	accessLevels := []AccessLevel{
		LvlOwner,
		LvlRead,
		LvlWrite,
		LvlReadDelete,
	}

	for _, x := range objtypes {
		accessMap[x] = make(map[Perm]map[AccessLevel]bool)
		for _, y := range perms {
			accessMap[x][y] = make(map[AccessLevel]bool)
			for _, z := range accessLevels {
				accessMap[x][y][z] = false
			}
		}
	}

	for _, lvl := range []AccessLevel{LvlRead, LvlWrite, LvlOwner, LvlReadDelete} {
		for tp := range accessMap {
			for perm := range accessMap[tp] {
				if lvl == LvlRead {
					if perm == List || perm == Read {
						accessMap[tp][perm][lvl] = true
					} else {
						accessMap[tp][perm][lvl] = false
					}
				}

				if lvl == LvlWrite {
					accessMap[tp][perm][lvl] = accessMap[tp][perm][LvlRead]
					if perm == Edit {
						accessMap[tp][perm][lvl] = true
					}
				}

				if lvl == LvlOwner {
					accessMap[tp][perm][lvl] = true
				}

				if lvl == LvlReadDelete {
					accessMap[tp][perm][lvl] = accessMap[tp][perm][LvlRead]
					if perm == Delete {
						accessMap[tp][perm][lvl] = true
					}
				}
			}
		}
	}

	if os.Getenv("GIN_MODE") != "release" {
		for k1 := range accessMap {
			for k2 := range accessMap[k1] {
				for k3 := range accessMap[k1][k2] {
					fmt.Printf("%s %s %s %t\n", k1, k2, k3, accessMap[k1][k2][k3])
				}
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
