package requests_tests

import (
	"testing"

	"git.containerum.net/ch/kube-client/pkg/cmd"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNamespace(test *testing.T) {
	client, err := cmd.CreateCmdClient(
		cmd.ClientConfig{
			ResourceAddr: "http://192.168.88.200:1213",
			APIurl:       "http://192.168.88.200:1214",
			User: cmd.User{
				Role: "admin",
			},
		})
	if err != nil {
		test.Fatalf("error while creating client: %v", err)
	}
	Convey("Test KubeAPI methods", test, func() {
		Convey("get namespace", func() {
			_, err := client.GetNamespace(kubeAPItestNamespace)
			So(err, ShouldBeNil)
		})
		Convey("get namespace list", func() {
			_, err := client.GetNamespaceList(map[string]string{})
			So(err, ShouldBeNil)
		})
	})
	Convey("Test resource service methods", test, func() {
		Convey("get namespace", func() {
			_, err := client.ResourceGetNamespace(resourceTestNamespace, "")
			So(err, ShouldBeNil)
		})
		Convey("get namespace list", func() {
			_, err := client.ResourceGetNamespaceList(0, 16, "")
			So(err, ShouldBeNil)
		})
	})
}
