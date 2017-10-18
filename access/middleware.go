package access

import (
	"fmt"
	"os"
	"sync"

	"bitbucket.org/exonch/kube-api/utils"

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
	LvlNone       AccessLevel = ""
)

// accessMap : objtype → Perm → role → (true|false)
var accessMap = make(map[string]map[Perm]map[AccessLevel]bool)
var accMapLock = &sync.RWMutex{}

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
					if tp == "Namespace" && (perm == Create || perm == Delete) {
						accessMap[tp][perm][lvl] = false
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

				if perm == Edit && tp == "Deployment" {
					accessMap[tp][perm][lvl] = false
				}
			}
		}
	}

	if os.Getenv("GIN_MODE") != "release" {
		fmt.Printf("kube-api/access.accessMap:\n")
		for k1 := range accessMap {
			for k2 := range accessMap[k1] {
				for k3 := range accessMap[k1][k2] {
					fmt.Printf("\t%s %s %s %t\n", k1, k2, k3, accessMap[k1][k2][k3])
				}
			}
		}
	}
}

// CheckAccess returns a gin.HandlerFunc that checks for specified access
// permissions in the supplied HTTP headers.
//
// The handler operates on "namespace" and "objectName" context values.
func CheckAccess(objtype string, perm Perm) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userdata *HTTPHeaders = c.MustGet("userData").(*HTTPHeaders)
		var alvl AccessLevel
		var verdict bool

		var ns2alvl = make(map[string]AccessLevel)
		for i, nsd := range userdata.Namespace {
			ns2alvl[nsd.ID] = nsd.Access
		}

		switch objtype {
		case "Namespace":
			switch perm {
			case List:

			}
		}
	}
}
