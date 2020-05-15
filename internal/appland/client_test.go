package appland

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestBuildUrl(t *testing.T) {
	client := MakeTestClient()
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
		Post("/api/api_keys").
		MatchHeader("Authorization", "Basic YWRtaW46YWRtaW4=").
		MatchType("json").
		Reply(200).
		JSON(map[string]string{"api_key": new_api_key})

	client := MakeTestClient()

	require.Nil(t, client.Login("admin", "admin"))
	assert.Equal(t, new_api_key, client.Context().APIKey)
}

func TestDeleteAPIKey(t *testing.T) {
	defer gock.Off()

	gock.New(url).
		Delete("/api/api_keys").
		MatchHeader("Authorization", "Bearer "+api_key).
		MatchType("json").
		Reply(200).
		JSON(map[string]string{})

	client := MakeTestClient()

	require.Nil(t, client.DeleteAPIKey())
	assert.Empty(t, client.Context().APIKey)
}

func TestTestAPIKeyNotFound(t *testing.T) {
	defer gock.Off()

	gock.New(url).
		Get("/api/scenarios/0").
		MatchHeader("Authorization", "Bearer "+api_key).
		MatchType("json").
		Reply(http.StatusNotFound)

	client := MakeTestClient()

	ok, err := client.TestAPIKey(api_key)
	require.True(t, ok)
	require.Nil(t, err)
}

func TestTestAPIKeyUnauthorized(t *testing.T) {
	defer gock.Off()

	gock.New(url).
		Get("/api/scenarios/0").
		MatchHeader("Authorization", "Bearer "+api_key).
		MatchType("json").
		Reply(http.StatusUnauthorized)

	client := MakeTestClient()

	ok, err := client.TestAPIKey(api_key)
	require.False(t, ok)
	require.Nil(t, err)
}

func TestTestAPIKeyNotOK(t *testing.T) {
	defer gock.Off()

	gock.New(url).
		Get("/api/scenarios/0").
		MatchHeader("Authorization", "Bearer "+api_key).
		MatchType("json").
		Reply(http.StatusInternalServerError)

	client := MakeTestClient()

	ok, err := client.TestAPIKey(api_key)
	require.False(t, ok)
	require.NotNil(t, err)
}

func TestTestAPIKeyOK(t *testing.T) {
	defer gock.Off()

	uuid := "not-a-real-scenario-id"
	gock.New(url).
		Get("/api/scenarios/0").
		MatchHeader("Authorization", "Bearer "+api_key).
		MatchType("json").
		Reply(http.StatusOK).
		JSON(map[string]string{"uuid": uuid})

	client := MakeTestClient()

	require.Panics(t, func() {
		client.TestAPIKey(api_key)
	})
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

	client := MakeTestClient()
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

	client := MakeTestClient()
	res, err := client.CreateMapSet("myapp", "myorg", scenarios)
	require.Nil(t, err)
	assert.Equal(t, uint32(12345), res.ID)
	assert.Equal(t, uint32(67890), res.AppID)
}
