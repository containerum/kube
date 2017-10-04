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

// type testResponseWriter struct {
// 	*bytes.Buffer
// }
//
// func (rw testResponseWriter) Header() http.Header {
// 	return map[string][]string{}
// }
//
// func (rw testResponseWriter) WriteHeader(n int) {
// 	log.WriteString(fmt.Sprint(n, "\n"))
// }

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

func TestNamespaceSubstitutionsInNamespaces(t *testing.T) {
	testdataraw, err := ioutil.ReadFile("testdata/test_substitutions.json")
	if err != nil {
		t.Fatalf("cannot open test data file: %T %[1]v", err)
	}

	var testdata testSubsType
	err = json.Unmarshal(testdataraw, &testdata)
	if err != nil {
		t.Fatalf("cannot parse test data file: %T %[1]v", err)
	}

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

		gotns := *c.MustGet("requestObject").(*v1.Namespace)
		var wantns v1.Namespace
		json.Unmarshal(nstest.Want.Body, &wantns)

		if wantns.Namespace != gotns.Namespace {
			t.Fatalf("mismatch in namespace %d", i)
			t.Fatalf("wanted %#v", wantns)
			t.Fatalf("got    %#v", gotns)
		}
	}

}

func TestNamespaceSubstitutionsInDeployments(t *testing.T) {
	testdataraw, err := ioutil.ReadFile("testdata/test_substitutions.json")
	if err != nil {
		t.Fatalf("cannot open test data file: %T %[1]v", err)
	}

	var testdata testSubsType
	err = json.Unmarshal(testdataraw, &testdata)
	if err != nil {
		t.Fatalf("cannot parse test data file: %T %[1]v", err)
	}

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
			t.Errorf("mismatch in deployment %d", i)
			t.Errorf("wanted %s", deplstr(want))
			t.Errorf("got    %s", deplstr(got ))
			t.FailNow()
		}
	}
}

func TestVolumeSubstitutionsInDeployments(t *testing.T) {
	testdataraw, err := ioutil.ReadFile("testdata/test_substitutions.json")
	if err != nil {
		t.Fatalf("cannot open test data file: %T %[1]v", err)
	}

	var testdata testSubsType
	err = json.Unmarshal(testdataraw, &testdata)
	if err != nil {
		t.Fatalf("cannot parse test data file: %T %[1]v", err)
	}

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

		//TODO
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
