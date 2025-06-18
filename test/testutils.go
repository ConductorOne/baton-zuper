package test

import (
	"context"
	"encoding/json"
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

// MockClient is a mock implementation of the Zuper client for testing.
type MockClient struct {
	GetUsersFunc             func(ctx context.Context, options client.PageOptions) ([]*client.ZuperUser, string, annotations.Annotations, error)
	GetUserByIDFunc          func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error)
	CreateUserFunc           func(ctx context.Context, user client.UserPayload) (*client.CreateUserResponse, annotations.Annotations, error)
	GetTeamsFunc             func(ctx context.Context, options client.PageOptions) ([]*client.Team, string, annotations.Annotations, error)
	GetTeamUsersFunc         func(ctx context.Context, teamID string) ([]*client.ZuperUser, string, annotations.Annotations, error)
	AssignUserToTeamFunc     func(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error)
	UnassignUserFromTeamFunc func(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error)
	UpdateUserRoleFunc       func(ctx context.Context, userUID string, roleID int) (*client.UpdateUserRoleResponse, annotations.Annotations, error)
}

// GetUsers calls the mock method if it is defined.
func (m *MockClient) GetUsers(ctx context.Context, options client.PageOptions) ([]*client.ZuperUser, string, annotations.Annotations, error) {
	if m.GetUsersFunc != nil {
		return m.GetUsersFunc(ctx, options)
	}
	return nil, "", nil, nil
}

// CreateUser calls the mock method if it is defined.
func (m *MockClient) CreateUser(ctx context.Context, user client.UserPayload) (*client.CreateUserResponse, annotations.Annotations, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil, nil, nil
}

// GetUserByID calls the mock method if it is defined.
func (m *MockClient) GetUserByID(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userUID)
	}
	return nil, nil, nil
}

// GetTeams calls the mock method if it is defined.
func (m *MockClient) GetTeams(ctx context.Context, options client.PageOptions) ([]*client.Team, string, annotations.Annotations, error) {
	if m.GetTeamsFunc != nil {
		return m.GetTeamsFunc(ctx, options)
	}
	return nil, "", nil, nil
}

// GetTeamUsers calls the mock method if it is defined.
func (m *MockClient) GetTeamUsers(ctx context.Context, teamID string) ([]*client.ZuperUser, string, annotations.Annotations, error) {
	if m.GetTeamUsersFunc != nil {
		return m.GetTeamUsersFunc(ctx, teamID)
	}
	return nil, "", nil, nil
}

// AssignUserToTeam calls the mock method if it is defined.
func (m *MockClient) AssignUserToTeam(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error) {
	if m.AssignUserToTeamFunc != nil {
		return m.AssignUserToTeamFunc(ctx, teamUID, userUID)
	}
	return nil, nil, nil
}

// UnassignUserFromTeam calls the mock method if it is defined.
func (m *MockClient) UnassignUserFromTeam(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error) {
	if m.UnassignUserFromTeamFunc != nil {
		return m.UnassignUserFromTeamFunc(ctx, teamUID, userUID)
	}
	return nil, nil, nil
}

// UpdateUserRole calls the mock method if it is defined.
func (m *MockClient) UpdateUserRole(ctx context.Context, userUID string, roleID int) (*client.UpdateUserRoleResponse, annotations.Annotations, error) {
	if m.UpdateUserRoleFunc != nil {
		return m.UpdateUserRoleFunc(ctx, userUID, roleID)
	}
	return nil, nil, nil
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

// LoadMockJSON loads the content of a mock JSON file from /test/mock/ as []byte.
func LoadMockJSON(fileName string) []byte {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	fullPath := filepath.Join(baseDir, "mock", fileName)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		panic(err)
	}
	return data
}

// LoadMockStruct loads a mock JSON file and unmarshals it into the provided interface.
func LoadMockStruct(fileName string, v interface{}) {
	data := LoadMockJSON(fileName)
	if err := json.Unmarshal(data, v); err != nil {
		panic(err)
	}
}

// ReadTeamsFile loads content from a JSON file from /test/mock/ for teams.
func ReadTeamsFile(fileName string) string {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	fullPath := filepath.Join(baseDir, "mock", fileName)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// LoadMockTeamsStruct loads a mock JSON file and unmarshals it into the provided interface for teams.
func LoadMockTeamsStruct(fileName string, v interface{}) {
	data := LoadMockJSON(fileName)
	if err := json.Unmarshal(data, v); err != nil {
		panic(err)
	}
}
