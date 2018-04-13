package middleware

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckUserRole(t *testing.T) {
	Convey("Check user role", t, func() {
		Convey("Check empty role", func() {
			ok, err := checkIsUserRole("")
			So(err, ShouldBeNil)
			So(ok, ShouldEqual, false)
		})
		Convey("Check admin role", func() {
			ok, err := checkIsUserRole(RoleAdmin)
			So(err, ShouldBeNil)
			So(ok, ShouldEqual, false)
		})
		Convey("Check user role", func() {
			ok, err := checkIsUserRole(RoleUser)
			So(err, ShouldBeNil)
			So(ok, ShouldEqual, true)
		})
		Convey("Check wrong role", func() {
			ok, err := checkIsUserRole("useradmin")
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrInvalidUserRole)
			So(ok, ShouldEqual, false)
		})
	})
}
