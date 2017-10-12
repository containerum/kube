package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	//meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gin-gonic/gin"
)

func init() {
	log.SetFlags(log.LstdFlags)
}

// testSubsType is for unmarshaling JSON test data onto it.
type testSubsType struct {
	Kube, User struct {
		Namespace, Deployment, DeploymentList, Service, ServiceList []struct {
			In, Want struct {
				HTTPHeaders map[string]string `json:"http-headers,omitempty"`
				Body        json.RawMessage
			}
		}
	}
}

var testdata testSubsType

func TestMain(m *testing.M) {
	testdataraw, err := ioutil.ReadFile("testdata/test_substitutions.json")
	if err != nil {
		log.Fatalf("cannot open test data file: %T %[1]v", err)
	}

	err = json.Unmarshal(testdataraw, &testdata)
	if err != nil {
		log.Fatalf("cannot parse test data file: %T %[1]v", err)
	}

	m.Run()
}

func TestNamespaceSubstitutionsInNamespaces(t *testing.T) {
	for i, nstest := range testdata.Kube.Namespace {
		c, _ := gin.CreateTestContext(nil)
		httphead := http.Header{}
		for k, v := range nstest.In.HTTPHeaders {
			httphead.Set(k, v)
		}
		c.Request = &http.Request{
			Header: httphead,
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(nstest.In.Body))),
		}

		ParseJSON(c)
		SubstitutionsFromHeadersFor("requestObject", false)(c)

		got := *c.MustGet("requestObject").(*v1.Namespace)
		var want v1.Namespace
		json.Unmarshal(nstest.Want.Body, &want)

		if want.Name != got.Name {
			t.Errorf("mismatch in namespace %d", i+1)
			t.Errorf("wanted metadata.namespace=%q", want.ObjectMeta.Name)
			t.Errorf("got    metadata.namespace=%q", got.ObjectMeta.Name)
			t.FailNow()
		}
	}

}

func TestNamespaceSubstitutionsInDeployments(t *testing.T) {
	for i, deptest := range testdata.Kube.Deployment {
		c, _ := gin.CreateTestContext(nil)
		httphead := http.Header{}
		for k, v := range deptest.In.HTTPHeaders {
			httphead.Set(k, v)
		}
		c.Request = &http.Request{
			Header: httphead,
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(deptest.In.Body))),
		}

		ParseJSON(c)
		SubstitutionsFromHeadersFor("requestObject", false)(c)

		got := *c.MustGet("requestObject").(*v1beta1.Deployment)
		var want v1beta1.Deployment
		json.Unmarshal(deptest.Want.Body, &want)

		if want.Namespace != got.Namespace {
			t.Errorf("mismatch in deployment %d", i+1)
			t.Errorf("wanted %s", deplstr(want))
			t.Errorf("got    %s", deplstr(got))
			t.FailNow()
		}
	}
}

func TestVolumeSubstitutionsInDeployments(t *testing.T) {
	for i, deptest := range testdata.Kube.Deployment {
		c, _ := gin.CreateTestContext(nil)
		httphead := http.Header{}
		for k, v := range deptest.In.HTTPHeaders {
			httphead.Set(k, v)
		}
		c.Request = &http.Request{
			Header: httphead,
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(deptest.In.Body))),
		}

		ParseJSON(c)
		SubstitutionsFromHeadersFor("requestObject", false)(c)

		got := *c.MustGet("requestObject").(*v1beta1.Deployment)
		var want v1beta1.Deployment
		json.Unmarshal(deptest.Want.Body, &want)

		for j, vol := range got.Spec.Template.Spec.Volumes {
			if vol.Name != want.Spec.Template.Spec.Volumes[j].Name {
				t.Fatalf("mismatch in deployment %d volume %d (got name %s, wanted %s)",
					i+1, j+1, vol.Name, want.Spec.Template.Spec.Volumes[j].Name)
			}
		}
	}
}

func deplstr(depl v1beta1.Deployment) string {
	var b, err = json.Marshal(depl)
	if err != nil {
		b, _ = json.Marshal(map[string]string{
			"error": err.Error(),
		})
	}
	return string(b)
}
