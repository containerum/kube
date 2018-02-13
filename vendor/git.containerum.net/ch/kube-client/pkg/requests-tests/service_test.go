package requests_tests

import (
	"testing"

	"git.containerum.net/ch/kube-client/pkg/cmd"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	kubeAPItestService = "ch-glusterfs"
)

func TestService(test *testing.T) {
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
	Convey("Test Kube API methods", test, func() {
		Convey("get service", func() {
			_, err := client.GetService(kubeAPItestNamespace, kubeAPItestService)
			So(err, ShouldBeNil)
		})
		Convey("get service list", func() {
			_, err := client.GetServiceList(kubeAPItestNamespace)
			So(err, ShouldBeNil)
		})
	})
}
