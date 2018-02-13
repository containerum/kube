package cmd

import (
	"net/url"
	"os"

	"github.com/go-resty/resty"
)

//TODO: Make Interface

//Client - rest client
type Client struct {
	*resty.Request
	serverURL           string
	resourceServiceAddr string
	User                User
}

//User -
type User struct {
	Role string
}

type ClientConfig struct {
	User         User
	APIurl       string
	ResourceAddr string
}

//CreateCmdClient -
func CreateCmdClient(config ClientConfig) (*Client, error) {
	var APIurl *url.URL
	var err error
	if config.APIurl == "" {
		APIurl, err = url.Parse(os.Getenv("API_URL"))
	} else {
		APIurl, err = url.Parse(config.APIurl)
	}
	if err != nil {
		return nil, err
	}
	config.APIurl = APIurl.String()

	if config.ResourceAddr == "" {
		// TODO: addr validation
		config.ResourceAddr = os.Getenv("RESOURCE_ADDR")
	}
	client := &Client{
		Request:             resty.R(),
		serverURL:           config.APIurl,
		resourceServiceAddr: config.ResourceAddr,
		User:                config.User,
	}
	client.SetHeaders(map[string]string{
		"X-User-Role": client.User.Role,
	})
	return client, nil
}
