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

// TestUserBuilder_List tests the List method of the userBuilder using mock responses.
// It validates correct parsing of user data, handling of pagination tokens, annotations, and error scenarios.
func TestUserBuilder_List(t *testing.T) {
	tests := []struct {
		name           string
		mockFile       string
		nextToken      string
		expectError    bool
		expectEmpty    bool
		expectedStatus func(user client.ZuperUser) v2.UserTrait_Status_Status
	}{
		{
			name:        "success with valid user data",
			mockFile:    "users_success.json",
			nextToken:   "",
			expectError: false,
			expectEmpty: false,
			expectedStatus: func(u client.ZuperUser) v2.UserTrait_Status_Status {
				if u.IsDeleted || !u.IsActive {
					return v2.UserTrait_Status_STATUS_DELETED
				}
				return v2.UserTrait_Status_STATUS_ENABLED
			},
		},
		{
			name:           "client error",
			mockFile:       "",
			nextToken:      "",
			expectError:    true,
			expectEmpty:    true,
			expectedStatus: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockUsers []*client.ZuperUser
			var annos annotations.Annotations

			if tt.mockFile != "" {
				mockData := test.ReadFile(tt.mockFile)
				err := json.Unmarshal([]byte(mockData), &mockUsers)
				require.NoError(t, err)
			}

			mockCli := &test.MockClient{
				GetUsersFunc: func(ctx context.Context, options client.PageOptions) ([]*client.ZuperUser, string, annotations.Annotations, error) {
					if tt.expectError {
						return nil, "", nil, errors.New("mock client error")
					}
					return mockUsers, tt.nextToken, annos, nil
				},
				GetUserByIDFunc: func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
					return nil, nil, nil // No se usa en estos tests
				},
			}

			builder := &userBuilder{
				resourceType: userResourceType,
				client:       mockCli,
			}

			resources, nextPage, gotAnnos, err := builder.List(context.Background(), nil, &pagination.Token{Token: ""})

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

			for i, r := range resources {
				expected := mockUsers[i]
				assert.Contains(t, r.DisplayName, expected.FirstName)
				assert.Contains(t, r.DisplayName, expected.LastName)
			}
		})
	}
}
