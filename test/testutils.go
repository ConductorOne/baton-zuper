package test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-zuper/pkg/client"
)

// Mock constants.
const (
	MockBaseURL     = "https://mock.api.zuper.co"
	MockAccessToken = "mock-access-token"
)

// MockClient is a mock implementation of the Zuper client.
type MockClient struct {
	GetUsersFunc func(ctx context.Context, token string) ([]*client.ZuperUser, string, annotations.Annotations, error)
}

// GetUsers calls the mock method if it is defined.
func (m *MockClient) GetUsers(ctx context.Context, token string) ([]*client.ZuperUser, string, annotations.Annotations, error) {
	if m.GetUsersFunc != nil {
		return m.GetUsersFunc(ctx, token)
	}
	return nil, "", nil, nil
}

// ReadFile loads content from a JSON file from /test/mock/.
func ReadFile(fileName string) string {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	fullPath := filepath.Join(baseDir, "mock", fileName)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// CreateMockResponseBody creates an io.ReadCloser with the contents of the file.
func CreateMockResponseBody(fileName string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(ReadFile(fileName)))
}
