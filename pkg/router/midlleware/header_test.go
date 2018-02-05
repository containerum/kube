package middleware

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckUserRole(t *testing.T) {
	Convey("Check user role", t, func() {
		Convey("Check empty role", func() {
			ok, err := checkUserRole("")
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrInvalidUserRole)
			So(ok, ShouldEqual, false)
		})
		Convey("Check admin role", func() {
			ok, err := checkUserRole("admin")
			So(err, ShouldBeNil)
			So(ok, ShouldEqual, true)
		})
		Convey("Check user role", func() {
			ok, err := checkUserRole("user")
			So(err, ShouldBeNil)
			So(ok, ShouldEqual, false)
		})
		Convey("Check wrong role", func() {
			ok, err := checkUserRole("useradmin")
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrInvalidUserRole)
			So(ok, ShouldEqual, false)
		})
	})
}
