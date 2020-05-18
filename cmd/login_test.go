package cmd

import (
	"strings"
	"testing"

	"github.com/applandinc/appland-cli/internal/appland"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	appland.Client
}

func (m *MockClient) Login(login string, password string) error {
	args := m.Called(login, password)
	return args.Error(0)
}

func TestLoginCommand(t *testing.T) {

	testClient := &MockClient{}

	testClient.On("Login", "bob", "secret").Return(nil)

	testConnecter := func() appland.Client {
		return appland.Client(testClient)
	}
	stdin := strings.NewReader("bob\n")
	passwordReader := func() ([]byte, error) {
		return []byte("secret\n"), nil
	}

	cmd := NewLoginCommand(testConnecter, stdin, passwordReader)

	cmd.Execute()

	testClient.AssertExpectations(t)
}
