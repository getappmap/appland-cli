package appland

import (
	"strings"
	"testing"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

var (
	url     = "http://example"
	api_key = "my_api_key"
)

func makeTestClient() *Client {
	return MakeClient(&config.Context{
		APIKey: api_key,
		URL:    url,
	})
}

func TestBuildUrl(t *testing.T) {
	client := makeTestClient()
	assert := assert.New(t)
	assert.Equal(client.BuildUrl(), "http://example")
	assert.Equal(client.BuildUrl("applications", 1), "http://example/applications/1")
	assert.Equal(client.BuildUrl("applications", 1, true, false), "http://example/applications/1/true/false")
	assert.Equal(client.BuildUrl(1.123), "http://example/1.123")
}

func TestLogin(t *testing.T) {
	defer gock.Off()

	new_api_key := "3977c1ca-6d87-49ae-bac7-97a9440a0149"
	gock.New(url).
		Post("/api/api_key").
		MatchHeader("Authorization", "Basic YWRtaW46YWRtaW4=").
		MatchType("json").
		Reply(200).
		JSON(map[string]string{"api_key": new_api_key})

	client := makeTestClient()

	require.Nil(t, client.Login("admin", "admin"))
	assert.Equal(t, new_api_key, client.context.APIKey)
}

func TestCreateScenario(t *testing.T) {
	defer gock.Off()

	scenarioUUID := "100582f6-27ba-4a04-a9d6-a634c742076c"

	gock.New(url).
		Post("/api/scenarios").
		MatchHeader("Authorization", "Bearer "+api_key).
		MatchType("json").
		JSON(map[string]string{"org": "myorg", "data": "{}"}).
		Reply(201).
		JSON(map[string]string{"uuid": scenarioUUID})

	client := makeTestClient()
	res, err := client.CreateScenario("myorg", strings.NewReader("{}"))
	require.Nil(t, err)
	assert.Equal(t, scenarioUUID, res.UUID)
}

func TestCreateMapSet(t *testing.T) {
	defer gock.Off()

	scenarios := []string{
		"100582f6-27ba-4a04-a9d6-a634c742076c",
		"e06eb6b2-1031-4625-8218-4b6a65580584",
	}

	gock.New(url).
		Post("/api/mapsets").
		MatchHeader("Authorization", "Bearer "+api_key).
		MatchType("json").
		JSON(map[string]interface{}{"org": "myorg", "app": "myapp", "scenarios": scenarios}).
		Reply(201).
		JSON(map[string]uint32{"id": 12345, "app_id": 67890})

	client := makeTestClient()
	res, err := client.CreateMapSet("myapp", "myorg", scenarios)
	require.Nil(t, err)
	assert.Equal(t, uint32(12345), res.ID)
	assert.Equal(t, uint32(67890), res.AppID)
}

func TestDeleteAPIKey(t *testing.T) {
	defer gock.Off()

	gock.New(url).
		Delete("/api/api_key").
		MatchHeader("Authorization", "Bearer "+api_key).
		MatchType("json").
		Reply(200).
		JSON(map[string]string{})

	// assert := assert.New(t)

	// client.Delete
}
