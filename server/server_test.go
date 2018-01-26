package server

import (
	"os"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
)

type testDataType struct {
	Namespaces  []*v1.Namespace
	Deployments []*v1beta1.Deployment
}

var testdata testDataType

func TestMain(m *testing.M) {
	LoadKubeClients(os.Getenv("CH_KUBE_API_KUBE_CONF"))
	//err := jsoniter.Unmarshal(testdataraw, &testdata)
	os.Setenv("GIN_MODE", "release")
	m.Run()
}

func Test001GetNamespaceBefore(t *testing.T) {}
func Test002CreateNamespace(t *testing.T)    {}
func Test003GetNamespace(t *testing.T)       {}
func Test004ListNamespaces(t *testing.T)     {}
func Test005DeleteNamespace(t *testing.T)    {}

func Test006GetServiceBefore(t *testing.T) {}
func Test007CreateService(t *testing.T)    {}
func Test008GetService(t *testing.T)       {}
func Test009ListServices(t *testing.T)     {}
func Test010DeleteService(t *testing.T)    {}

func Test011GetSecretBefore(t *testing.T) {}
func Test012CreateSecret(t *testing.T)    {}
func Test013GetSecret(t *testing.T)       {}
func Test014ListSecrets(t *testing.T)     {}
func Test015DeleteSecret(t *testing.T)    {}

func Test016GetEndpointsBefore(t *testing.T) {}
func Test017CreateEndpoints(t *testing.T)    {}
func Test018GetEndpoints(t *testing.T)       {}
func Test019ListEndpoints(t *testing.T)      {}
func Test020DeleteEndpoints(t *testing.T)    {}

func Test021GetConfigMapBefore(t *testing.T) {}
func Test022CreateConfigMap(t *testing.T)    {}
func Test023GetConfigMap(t *testing.T)       {}
func Test024ListConfigMaps(t *testing.T)     {}
func Test025DeleteConfigMap(t *testing.T)    {}

func Test026GetDeploymentBefore(t *testing.T) {}
func Test027CreateDeployment(t *testing.T)    {}
func Test028GetDeployment(t *testing.T)       {}
func Test029ListDeployments(t *testing.T)     {}
func Test030DeleteDeployment(t *testing.T)    {}
