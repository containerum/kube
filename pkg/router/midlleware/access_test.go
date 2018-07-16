package middleware

import (
	"net/http"
	"testing"

	"github.com/appleboy/gofight"
	kubeModel "github.com/containerum/kube-client/pkg/model"
	headers "github.com/containerum/utils/httputil"
	"github.com/gin-gonic/gin"

	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestContainsAccess(t *testing.T) {
	Convey("Test containsAccess func", t, func() {
		Convey("Check empty access in empty list", func() {
			So(containsAccess(""), ShouldBeFalse)
		})
		Convey("Check empty access in read levels array", func() {
			So(containsAccess("", readLevels...), ShouldBeFalse)
		})
		Convey("Check read access in empty list", func() {
			So(containsAccess("read"), ShouldBeFalse)
		})
		Convey("Check read access in list without read access", func() {
			So(containsAccess("read", kubeModel.Owner, kubeModel.Write), ShouldBeFalse)
		})
		Convey("Check read access in read levels array", func() {
			So(containsAccess("read", readLevels...), ShouldBeTrue)
		})
	})
}

func TestIsAdmin(t *testing.T) {
	e := gin.New()
	r := gofight.New()
	e.GET("/test", IsAdmin, func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	})
	Convey("Test IsAdmin Middleware", t, func() {
		Convey("Check without User-Role", func() {
			r.GET("/test").
				Run(e, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
					So(r.Code, ShouldEqual, http.StatusForbidden)
				})
		})
		Convey("Check wrong User-Role", func() {
			r.GET("/test").
				SetHeader(gofight.H{
					headers.UserRoleXHeader: "useradmin",
				}).
				Run(e, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
					So(r.Code, ShouldEqual, http.StatusForbidden)
				})
		})
		Convey("Check user User-Role", func() {
			r.GET("/test").
				SetHeader(gofight.H{
					headers.UserRoleXHeader: RoleUser,
				}).
				Run(e, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
					So(r.Code, ShouldEqual, http.StatusForbidden)
				})
		})
		Convey("Check admin User-Role", func() {
			r.GET("/test").
				SetHeader(gofight.H{
					headers.UserRoleXHeader: RoleAdmin,
				}).
				Run(e, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
					So(r.Code, ShouldEqual, http.StatusOK)
				})
		})
	})
}
