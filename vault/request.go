package vault

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/flaccid/vsync/config"
)

// Request performs a request with a vault client
func (v *Client) Request(appConfig *config.AppConfig, destinationVault bool, method, uri string, body interface{}) (*http.Response, error) {
	if destinationVault {
		client = appConfig.Destination.Client
	} else {
		client = appConfig.Source.Client
	}

	url := fmt.Sprintf("/%s/%s", apiVersion, strings.TrimPrefix(uri, "/"))
	log.Debugf("make request: %s %s, body: %#v", method, url, body)

	request := client.NewRequest(method, url)
	if err := request.SetJSONBody(body); err != nil {
		return nil, err
	}

	resp, err := client.RawRequest(request)
	if err != nil {
		return nil, err
	}

	return resp.Response, nil
}
