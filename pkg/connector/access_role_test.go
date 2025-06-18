package connector

import (
	"context"
	"errors"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-zuper/pkg/client"
	"github.com/conductorone/baton-zuper/test"
	"github.com/stretchr/testify/assert"
)

func TestAccessRoleBuilder_Grant(t *testing.T) {
	mockUser := &client.ZuperUser{
		UserUID:    "user-1",
		FirstName:  "Test",
		LastName:   "User",
		Email:      "test@example.com",
		AccessRole: &client.AccessRole{AccessRoleUID: "role-1", AccessRoleName: "Role 1"},
	}
	mockCli := &test.MockClient{
		GetUserByIDFunc: func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
			return mockUser, nil, nil
		},
		UpdateUserAccessRoleFunc: func(ctx context.Context, userUID string, accessRoleUID string) (*client.UpdateUserRoleResponse, annotations.Annotations, error) {
			return &client.UpdateUserRoleResponse{Message: "Access role updated"}, nil, nil
		},
	}
	builder := &accessRoleBuilder{
		resourceType: accessRoleResourceType,
		client:       mockCli,
	}
	accessRoleRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: accessRoleResourceType.Id,
			Resource:     "role-2",
		},
		DisplayName: "Role 2",
	}
	userRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: userResourceType.Id,
			Resource:     "user-1",
		},
	}
	ent := &v2.Entitlement{
		Resource: accessRoleRes,
	}

	t.Run("assigns access role if user does not have it", func(t *testing.T) {
		mockUser.AccessRole = &client.AccessRole{AccessRoleUID: "role-1"}
		ent.Resource.Id.Resource = "role-2"
		grants, annos, err := builder.Grant(context.Background(), userRes, ent)
		assert.NoError(t, err)
		assert.NotNil(t, grants)
		assert.Equal(t, 0, len(annos))
	})

	t.Run("does nothing if user already has access role", func(t *testing.T) {
		mockUser.AccessRole = &client.AccessRole{AccessRoleUID: "role-2"}
		ent.Resource.Id.Resource = "role-2"
		grants, annos, err := builder.Grant(context.Background(), userRes, ent)
		assert.NoError(t, err)
		assert.Nil(t, grants)
		assert.NotNil(t, annos)
	})

	t.Run("returns error if client fails", func(t *testing.T) {
		mockCli.GetUserByIDFunc = func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
			return nil, nil, errors.New("mock error")
		}
		grants, annos, err := builder.Grant(context.Background(), userRes, ent)
		assert.Error(t, err)
		assert.Nil(t, grants)
		assert.Nil(t, annos)
	})
}

func TestAccessRoleBuilder_Revoke(t *testing.T) {
	called := false
	mockUserWithAccessRole := &client.ZuperUser{
		UserUID:    "user-1",
		FirstName:  "Test",
		LastName:   "User",
		Email:      "test@example.com",
		AccessRole: &client.AccessRole{AccessRoleUID: "role-1", AccessRoleName: "Role 1"},
	}
	mockUserWithoutAccessRole := &client.ZuperUser{
		UserUID:    "user-1",
		FirstName:  "Test",
		LastName:   "User",
		Email:      "test@example.com",
		AccessRole: nil,
	}
	mockCli := &test.MockClient{
		UpdateUserAccessRoleFunc: func(ctx context.Context, userUID string, accessRoleUID string) (*client.UpdateUserRoleResponse, annotations.Annotations, error) {
			called = true
			assert.Equal(t, "user-1", userUID)
			assert.Equal(t, "", accessRoleUID)
			return &client.UpdateUserRoleResponse{Message: "Access role removed"}, nil, nil
		},
	}
	builder := &accessRoleBuilder{
		resourceType: accessRoleResourceType,
		client:       mockCli,
	}
	grant := &v2.Grant{
		Principal: &v2.Resource{Id: &v2.ResourceId{ResourceType: userResourceType.Id, Resource: "user-1"}},
	}

	t.Run("always revokes access role (idempotent)", func(t *testing.T) {
		called = false
		mockCli.GetUserByIDFunc = func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
			return mockUserWithAccessRole, nil, nil
		}
		annos, err := builder.Revoke(context.Background(), grant)
		assert.NoError(t, err)
		assert.True(t, called)
		assert.Nil(t, annos)
	})

	t.Run("returns GrantAlreadyRevoked if user has no access role", func(t *testing.T) {
		mockCli.GetUserByIDFunc = func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
			return mockUserWithoutAccessRole, nil, nil
		}
		annos, err := builder.Revoke(context.Background(), grant)
		assert.NoError(t, err)
		assert.NotNil(t, annos)
	})

	t.Run("returns error if client fails", func(t *testing.T) {
		mockCli.GetUserByIDFunc = func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
			return nil, nil, errors.New("mock error")
		}
		annos, err := builder.Revoke(context.Background(), grant)
		assert.Error(t, err)
		assert.Nil(t, annos)
	})
}
