package access

import (
	"fmt"
	"os"
	"sync"

	"bitbucket.org/exonch/kube-api/model"

	"k8s.io/api/apps/v1beta1"

	"github.com/gin-gonic/gin"
)

type HTTPHeaders struct {
	Namespace []model.UserData
	Volume    []model.UserData
}

func (hh *HTTPHeaders) NamespaceAccess(nsname string) AccessLevel {
	for _, ud := range hh.Namespace {
		if ud.ID == nsname {
			return AccessLevel(ud.Access)
		}
	}
	return AccessLevel("")
}

func (hh *HTTPHeaders) VolumeAccess(volname string) AccessLevel {
	for _, ud := range hh.Volume {
		if ud.ID == volname {
			return AccessLevel(ud.Access)
		}
	}
	return AccessLevel("")
}

// Required permission for the action on the object.
type Perm string

const (
	Create      Perm = "create"
	Delete      Perm = "delete"
	List        Perm = "list"
	Read        Perm = "read"
	Edit        Perm = "edit"
	SetImage    Perm = "setimage"
	SetReplicas Perm = "setreplicas"
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

// accessMap : m[objtype] → m[Perm] → m[role] → true/false
var accessMap = make(map[string]map[Perm]map[AccessLevel]bool)
var accMapLock = &sync.RWMutex{}

func init() {
	objtypes := []string{
		"ConfigMap",
		"Deployment",
		"Endpoints",
		"Ingress",
		"Namespace",
		"Secret",
		"Service",
	}
	perms := []Perm{
		Create,
		Delete,
		List,
		Read,
		Edit,
		SetImage,
		SetReplicas,
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
					if perm == Edit || perm == SetImage || perm == SetReplicas {
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

	{
		aMapVol := make(map[Perm]map[AccessLevel]bool)
		accessMap["Volume"] = aMapVol
		aMapVol[Read] = make(map[AccessLevel]bool)
		aMapVol[Edit] = make(map[AccessLevel]bool)

		aMapVol[Read][LvlOwner] = true
		aMapVol[Edit][LvlOwner] = true

		aMapVol[Read][LvlRead] = true
		aMapVol[Edit][LvlRead] = false

		aMapVol[Read][LvlWrite] = true
		aMapVol[Edit][LvlWrite] = true

		aMapVol[Read][LvlReadDelete] = true
		aMapVol[Edit][LvlReadDelete] = false
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
// Must be used AFTER user data substitutions.
func CheckAccess(objtype string, perm Perm) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userdata *HTTPHeaders = c.MustGet("userData").(*HTTPHeaders)
		accMapLock.Lock()
		var canAccess map[AccessLevel]bool = accessMap[objtype][perm]
		defer accMapLock.Unlock()
		var nsname, objname string
		var verdict bool

		if objtype == "Namespace" && perm == List {
			return
		}

		nsname = c.MustGet("namespace").(string)
		verdict = canAccess[userdata.NamespaceAccess(nsname)]
		if !verdict {
			accessDenied(c, "namespace rights")
			return
		}

		if objtype == "Deployment" && (perm == Edit || perm == Create) { // also check volumes
			var obj interface{}
			var ok, exists bool
			var depl *v1beta1.Deployment
			var containerPerms = make(map[string]Perm) // m[volumeName] → Read/Edit

			obj, exists = c.Get("requestObject")
			if !exists {
				goto endvol
			}
			depl, ok = obj.(*v1beta1.Deployment)
			if !ok {
				goto endvol
			}

			for _, cont := range depl.Spec.Template.Spec.Containers {
				for _, volmnt := range cont.VolumeMounts {
					_, exists := containerPerms[volmnt.Name]
					if !exists {
						containerPerms[volmnt.Name] = Read
					}
					if !volmnt.ReadOnly {
						containerPerms[volmnt.Name] = Edit
					}
				}
			}
			for vol, perm := range containerPerms {
				verdict = accessMap["Volume"][perm][userdata.VolumeAccess(vol)]
				if !verdict {
					if tmp, ok := c.Get("objectName"); ok {
						objname = tmp.(string)
					}
					accessDenied(c, "volume \""+vol+"\" in deployment \""+objname+"\"")
					return
				}
			}
		endvol:
		}

		if !verdict {
			accessDenied(c, "no info")
		}
	}
}

func accessDenied(c *gin.Context, ctxinfo string) {
	_, already := c.Get("accessDenied-mark")
	if already {
		return
	}
	c.Set("accessDenied-mark", true)
	c.AbortWithStatusJSON(401, map[string]string{
		"error":   "unauthorized (" + ctxinfo + ")",
		"errcode": "PERMISSION_DENIED",
	})
}
