package appland

import (
	"github.com/applandinc/appland-cli/internal/config"
)

var (
	url     = "http://example"
	api_key = "my_api_key"
)

func makeTestClient() Client {
	return MakeClient(&config.Context{
		APIKey: api_key,
		URL:    url,
	})
}

func MakeTestClient() Client {
	return makeTestClient()
}
