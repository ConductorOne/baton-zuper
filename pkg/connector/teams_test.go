package connector

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-zuper/pkg/client"
	"github.com/conductorone/baton-zuper/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTeamBuilder_List tests the List method of teamBuilder for correct team parsing, pagination, and error handling.
func TestTeamBuilder_List(t *testing.T) {
	tests := []struct {
		name        string
		mockFile    string
		nextToken   string
		expectError bool
		expectEmpty bool
	}{
		{
			name:        "success with valid team data",
			mockFile:    "teams_success.json",
			nextToken:   "",
			expectError: false,
			expectEmpty: false,
		},
		{
			name:        "client error",
			mockFile:    "",
			nextToken:   "",
			expectError: true,
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockTeams []*client.Team
			var annos annotations.Annotations

			if tt.mockFile != "" {
				mockData := test.ReadTeamsFile(tt.mockFile)
				err := json.Unmarshal([]byte(mockData), &mockTeams)
				require.NoError(t, err)
			}

			mockCli := &test.MockClient{
				GetTeamsFunc: func(ctx context.Context, options client.PageOptions) ([]*client.Team, string, annotations.Annotations, error) {
					if tt.expectError {
						return nil, "", nil, errors.New("mock client error")
					}
					return mockTeams, tt.nextToken, annos, nil
				},
				AssignUserToTeamFunc: func(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error) {
					return &client.AssignUserToTeamResponse{
						Type:    "success",
						Title:   "User assigned to team",
						Message: "User assigned to team successfully",
					}, nil, nil
				},
				UnassignUserFromTeamFunc: func(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error) {
					return &client.AssignUserToTeamResponse{
						Type:    "success",
						Title:   "User unassigned from team",
						Message: "User unassigned from team successfully",
					}, nil, nil
				},
			}

			builder := &teamBuilder{
				resourceType: teamResourceType,
				client:       mockCli,
			}

			resources, nextPage, gotAnnos, err := builder.List(context.Background(), nil, &pagination.Token{Token: "", Size: 50})

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, resources)
				return
			}

			require.NoError(t, err)

			if tt.expectEmpty {
				assert.Empty(t, resources)
				return
			}

			require.NotEmpty(t, resources)
			assert.Equal(t, tt.nextToken, nextPage)
			assert.Equal(t, len(annos), len(gotAnnos))
		})
	}
}

// TestTeamBuilder_Grants tests the Grants method of teamBuilder to ensure grants are created for team users.
func TestTeamBuilder_Grants(t *testing.T) {
	mockUsers := []*client.ZuperUser{{
		UserUID:   "user-1",
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@example.com",
	}}
	mockCli := &test.MockClient{
		GetTeamUsersFunc: func(ctx context.Context, teamID string) ([]*client.ZuperUser, string, annotations.Annotations, error) {
			return mockUsers, "", nil, nil
		},
	}
	builder := &teamBuilder{
		resourceType: teamResourceType,
		client:       mockCli,
	}
	teamRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: teamResourceType.Id,
			Resource:     "team-1",
		},
		DisplayName: "Team One",
	}
	grants, _, annos, err := builder.Grants(context.Background(), teamRes, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, grants)
	assert.Equal(t, "user-1", grants[0].Principal.Id.Resource)
	assert.Equal(t, "team-1", grants[0].Entitlement.Resource.Id.Resource)
	assert.Equal(t, 0, len(annos))
}

// TestTeamBuilder_Grants_Error tests the Grants method of teamBuilder for error handling when fetching team users fails.
func TestTeamBuilder_Grants_Error(t *testing.T) {
	mockCli := &test.MockClient{
		GetTeamUsersFunc: func(ctx context.Context, teamID string) ([]*client.ZuperUser, string, annotations.Annotations, error) {
			return nil, "", nil, errors.New("mock error")
		},
	}
	builder := &teamBuilder{
		resourceType: teamResourceType,
		client:       mockCli,
	}
	teamRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: teamResourceType.Id,
			Resource:     "team-err",
		},
		DisplayName: "Team Error",
	}
	grants, _, annos, err := builder.Grants(context.Background(), teamRes, nil)
	assert.Error(t, err)
	assert.Nil(t, grants)
	assert.Equal(t, 0, len(annos))
}

// TestTeamBuilder_Grant tests the Grant method of teamBuilder for assigning a user to a team.
func TestTeamBuilder_Grant(t *testing.T) {
	mockCli := &test.MockClient{
		AssignUserToTeamFunc: func(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error) {
			return &client.AssignUserToTeamResponse{Message: "User assigned to team"}, nil, nil
		},
	}
	builder := &teamBuilder{
		resourceType: teamResourceType,
		client:       mockCli,
	}
	teamRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: teamResourceType.Id,
			Resource:     "team-1",
		},
		DisplayName: "Team One",
	}
	userRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: userResourceType.Id,
			Resource:     "user-1",
		},
	}
	ent := &v2.Entitlement{
		Resource: teamRes,
	}
	grants, annos, err := builder.Grant(context.Background(), userRes, ent)
	assert.NoError(t, err)
	assert.NotNil(t, grants)
	assert.Equal(t, 0, len(annos))
}

// TestTeamBuilder_Grant_Error tests error handling in Grant method of teamBuilder.
func TestTeamBuilder_Grant_Error(t *testing.T) {
	mockCli := &test.MockClient{
		AssignUserToTeamFunc: func(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error) {
			return nil, nil, errors.New("mock assign error")
		},
	}
	builder := &teamBuilder{
		resourceType: teamResourceType,
		client:       mockCli,
	}
	teamRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: teamResourceType.Id,
			Resource:     "team-1",
		},
		DisplayName: "Team One",
	}
	userRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: userResourceType.Id,
			Resource:     "user-1",
		},
	}
	ent := &v2.Entitlement{
		Resource: teamRes,
	}
	grants, annos, err := builder.Grant(context.Background(), userRes, ent)
	assert.Error(t, err)
	assert.Nil(t, grants)
	assert.Nil(t, annos)
}

// TestTeamBuilder_Revoke tests the Revoke method of teamBuilder for unassigning a user from a team.
func TestTeamBuilder_Revoke(t *testing.T) {
	mockCli := &test.MockClient{
		UnassignUserFromTeamFunc: func(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error) {
			return &client.AssignUserToTeamResponse{Message: "User unassigned from team"}, nil, nil
		},
	}
	builder := &teamBuilder{
		resourceType: teamResourceType,
		client:       mockCli,
	}
	grant := &v2.Grant{
		Principal:   &v2.Resource{Id: &v2.ResourceId{ResourceType: userResourceType.Id, Resource: "user-1"}},
		Entitlement: &v2.Entitlement{Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: teamResourceType.Id, Resource: "team-1"}}},
	}
	annos, err := builder.Revoke(context.Background(), grant)
	assert.NoError(t, err)
	assert.Nil(t, annos)
}

// TestTeamBuilder_Revoke_Error tests error handling in Revoke method of teamBuilder.
func TestTeamBuilder_Revoke_Error(t *testing.T) {
	mockCli := &test.MockClient{
		UnassignUserFromTeamFunc: func(ctx context.Context, teamUID, userUID string) (*client.AssignUserToTeamResponse, annotations.Annotations, error) {
			return nil, nil, errors.New("mock unassign error")
		},
	}
	builder := &teamBuilder{
		resourceType: teamResourceType,
		client:       mockCli,
	}
	grant := &v2.Grant{
		Principal:   &v2.Resource{Id: &v2.ResourceId{ResourceType: userResourceType.Id, Resource: "user-1"}},
		Entitlement: &v2.Entitlement{Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: teamResourceType.Id, Resource: "team-1"}}},
	}
	annos, err := builder.Revoke(context.Background(), grant)
	assert.Error(t, err)
	assert.Nil(t, annos)
}
