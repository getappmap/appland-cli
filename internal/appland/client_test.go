package appland

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"regexp"
	"strings"
	"testing"

	"github.com/applandinc/appland-cli/internal/metadata"
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

type multipartPart struct {
	header textproto.MIMEHeader
	body   string
}

type multipartMatcher struct {
	*gock.MockMatcher
	parts []multipartPart
}

func (m *multipartMatcher) matchRequest(req *http.Request) (bool, error) {
	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if !(err == nil && strings.HasPrefix(mediaType, "multipart/")) {
		return false, err
	}

	body, err := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	mr := multipart.NewReader(ioutil.NopCloser(bytes.NewReader(body)), params["boundary"])
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		slurp, err := ioutil.ReadAll(p)
		if err != nil {
			return false, err
		}

		m.parts = append(m.parts, multipartPart{p.Header, string(slurp)})
	}

	return true, nil
}

func newMultipartMatcher() *multipartMatcher {
	m := multipartMatcher{gock.NewBasicMatcher(), nil}

	m.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		return m.matchRequest(req)
	})
	return &m
}

// tries to match full string or as a regex
func fuzzyMatch(needle string, haystack string) bool {
	if needle == haystack {
		return true
	}
	match, _ := regexp.MatchString(needle, haystack)
	return match
}

func headersMatch(needle textproto.MIMEHeader, haystack textproto.MIMEHeader) bool {
	for k := range needle {
		if !fuzzyMatch(needle.Get(k), haystack.Get(k)) {
			return false
		}
	}

	return true
}

func (m *multipartMatcher) matchPart(header textproto.MIMEHeader, body string) *multipartMatcher {
	m.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		for _, part := range m.parts {
			if headersMatch(header, part.header) {
				return fuzzyMatch(body, part.body), nil
			}
		}
		return false, nil
	})
	return m
}

func TestCreateScenario(t *testing.T) {
	defer gock.Off()

	scenarioUUID := "100582f6-27ba-4a04-a9d6-a634c742076c"

	matcher := newMultipartMatcher()
	matcher.matchPart(textproto.MIMEHeader{
		"Content-Disposition": {"attachment"},
		"Content-Type":        {"application/json"},
	}, "{}")
	matcher.matchPart(textproto.MIMEHeader{
		"Content-Disposition": {"inline"},
		"Content-Type":        {"application/json"},
	}, `{ "app": "myapp" }`)

	gock.New(url).
		Post("/api/scenarios").
		SetMatcher(matcher).
		MatchHeader("Authorization", "Bearer "+api_key).
		Reply(201).
		JSON(map[string]string{"uuid": scenarioUUID})

	client := MakeTestClient()
	res, err := client.CreateScenario("myapp", strings.NewReader("{}"))
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}
	require.Nil(t, err)
	assert.Equal(t, scenarioUUID, res.UUID)
}

func TestCreateMapSet(t *testing.T) {
	defer gock.Off()

	git := &metadata.Git{
		Commit: "76c0ae55fff17ae52ab67a0ff61e1af3d1157555",
		Branch: "master",
	}
	scenarios := []string{
		"100582f6-27ba-4a04-a9d6-a634c742076c",
		"e06eb6b2-1031-4625-8218-4b6a65580584",
	}

	gock.New(url).
		Post("/api/mapsets").
		MatchHeader("Authorization", "Bearer "+api_key).
		MatchType("json").
		JSON(
			map[string]interface{}{
				"app":       "myorg/myapp",
				"commit":    git.Commit,
				"branch":    git.Branch,
				"scenarios": scenarios}).
		Reply(201).
		JSON(map[string]uint32{"id": 12345, "app_id": 67890})

	client := MakeTestClient()
	mapset := BuildMapSet("myorg/myapp", scenarios).
		WithGitMetadata(git)

	res, err := client.CreateMapSet(mapset)
	require.Nil(t, err)
	assert.Equal(t, uint32(12345), res.ID)
	assert.Equal(t, uint32(67890), res.AppID)
}
