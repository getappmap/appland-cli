package recording

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockRecording struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var GetDoFunc func(req *http.Request) (*http.Response, error)

var (
	validAppmap = `
  {
    "metadata": {},
    "classMap": []
  }
  `
	invalidAppmap = "invalid json"
	checkEnabled  = `
  {
    "enabled": true
  }
  `
	checkDisabled = `
  {
    "enabled": false
  }
  `
)

func (m *MockRecording) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func init() {
	Client = &MockRecording{}
}

// START - status code 409 -> should return false
func TestStartRecordingInProgress(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte("")))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 409,
			Body:       body,
		}, nil
	}
	resp, err := StartRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, false, resp)
}

// START - status code 200 -> should return true
func TestStartNoRecordingInProgress(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte("")))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       body,
		}, nil
	}
	resp, err := StartRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, true, resp)
}

// START - unknown status code -> should return false and an error
func TestStartRecordingBadResponse(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte("")))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500, // unhandled status code
			Body:       body,
		}, nil
	}
	resp, err := StartRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, false, resp)
}

// STOP - status code 200 -> should return true
func TestStopRecordingInProgress(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte(validAppmap)))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       body,
		}, nil
	}
	resp, err := StopRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, true, resp)
}

// STOP - status code 404 -> should return false
func TestStopNoRecordingInProgress(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte("")))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 404,
			Body:       body,
		}, nil
	}
	resp, err := StopRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, false, resp)
}

// STOP - unknown status code -> should return false and an error
func TestStopBadStatusResponse(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte("")))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Body:       body,
		}, nil
	}
	resp, err := StopRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, false, resp)
}

// STOP - unable to read body -> should return false and an error
func TestStopBadBodyResponse(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte(invalidAppmap)))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Body:       body,
		}, nil
	}
	resp, err := StopRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, false, resp)
}

// CHECK - status code 200 and recording enabled -> should return true
func TestCheckRecordingInProgress(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte(checkEnabled)))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       body,
		}, nil
	}
	resp, err := StartRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, true, resp)
}

// CHECK - status code 200 and recording disabled -> should return false
func TestCheckNoRecordingInProgress(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte(checkDisabled)))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       body,
		}, nil
	}
	resp, err := CheckRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, false, resp)
}

// CHECK - unknown status code -> should return false and an error
func TestCheckBadStatusResponse(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewReader([]byte("")))
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Body:       body,
		}, nil
	}
	resp, err := StopRecording("url")
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.EqualValues(t, false, resp)
}
