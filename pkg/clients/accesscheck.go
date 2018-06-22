package clients

import (
	"context"

	"time"

	"git.containerum.net/ch/kube-api/pkg/kubeErrors"
	"github.com/containerum/cherry"
	"github.com/containerum/utils/httputil"
	"github.com/go-resty/resty"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type permissionsClient struct {
	client *resty.Client
}

func NewPermissionsAccessHTTPClient(host string) httputil.AccessChecker {
	log := logrus.WithField("component", "permissions-access")
	client := resty.New().
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetHostURL(host).
		SetDebug(true).
		SetTimeout(3*time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetError(cherry.Err{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return httputil.AccessChecker{
		PermissionsClient: permissionsClient{
			client: client,
		},
		AccessError:   kubeErrors.ErrAccessError,
		NotFoundError: kubeErrors.ErrResourceNotExist,
	}
}

func (client permissionsClient) GetAllAccesses(ctx context.Context) ([]httputil.ProjectAccess, error) {
	accesses := make([]httputil.ProjectAccess, 0)
	resp, err := client.client.R().
		SetContext(ctx).
		SetResult(&accesses).
		SetHeaders(httputil.RequestHeadersMap(ctx)).
		Get("/accesses")
	if err != nil {
		return accesses, err
	}
	if resp.Error() != nil {
		return accesses, resp.Error().(*cherry.Err)
	}
	return accesses, nil
}

func (client permissionsClient) GetNamespaceAccess(ctx context.Context, projectID, namespaceID string) (httputil.NamespaceAccess, error) {
	var access httputil.NamespaceAccess
	resp, err := client.client.R().
		SetContext(ctx).
		SetResult(&access).
		SetHeaders(httputil.RequestHeadersMap(ctx)).
		SetPathParams(map[string]string{
			"project":   projectID,
			"namespace": namespaceID,
		}).
		Get("/projects/{project}/projects/{project}/namespaces/{namespace}/access")
	if err != nil {
		return access, err
	}
	if resp.Error() != nil {
		return access, resp.Error().(*cherry.Err)
	}
	return access, nil
}
