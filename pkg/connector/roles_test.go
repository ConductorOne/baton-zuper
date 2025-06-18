package connector

import (
	"context"
	"errors"
	"strconv"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-zuper/pkg/client"
	"github.com/conductorone/baton-zuper/test"
	"github.com/stretchr/testify/assert"
)

// For testing, we create a roleBuilderTest that accepts the necessary interface.
type roleBuilderTest struct {
	resourceType *v2.ResourceType
	client       interface {
		GetUserByID(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error)
		UpdateUserRole(ctx context.Context, userUID string, roleID int) (*client.UpdateUserRoleResponse, annotations.Annotations, error)
	}
}

func (r *roleBuilderTest) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
	userID := principal.Id.Resource
	roleKey := entitlement.Resource.Id.Resource

	var roleIDStr string
	for _, definition := range roleDefinitions {
		if definition.RoleKey == roleKey {
			roleIDStr = definition.ID
			break
		}
	}
	if roleIDStr == "" {
		return nil, nil, errors.New("role ID not found for key: " + roleKey)
	}
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		return nil, nil, errors.New("invalid role ID: " + roleIDStr)
	}

	user, _, err := r.client.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if user.Role != nil && user.Role.RoleUID == roleIDStr {
		return nil, nil, nil
	}

	_, annos, err := r.client.UpdateUserRole(ctx, userID, roleID)
	if err != nil {
		return nil, annos, err
	}

	grantObj := &v2.Grant{
		Entitlement: entitlement,
		Principal:   principal,
	}
	return []*v2.Grant{grantObj}, annos, nil
}

func (r *roleBuilderTest) Revoke(ctx context.Context, g *v2.Grant) (annotations.Annotations, error) {
	userID := g.Principal.Id.Resource
	roleKey := g.Entitlement.Resource.Id.Resource

	var roleIDStr string
	for _, def := range roleDefinitions {
		if def.RoleKey == roleKey {
			roleIDStr = def.ID
			break
		}
	}
	if roleIDStr == "" {
		return nil, errors.New("role ID not found for key: " + roleKey)
	}

	user, _, err := r.client.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Role.RoleKey == "FIELD_EXECUTIVE" {
		return nil, nil
	}

	defaultRoleID := 3
	_, annos, err := r.client.UpdateUserRole(ctx, userID, defaultRoleID)
	if err != nil {
		return annos, err
	}
	return annos, nil
}

// TestRoleBuilder_Grant tests the Grant method of roleBuilderTest to ensure a role is assigned to a user.
func TestRoleBuilder_Grant(t *testing.T) {
	mockUser := &client.ZuperUser{
		UserUID:   "user-1",
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		Role:      &client.Role{RoleUID: "2", RoleKey: "TEAM_LEADER", RoleName: "Team Leader"},
	}
	mockCli := &test.MockClient{
		GetUserByIDFunc: func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
			return mockUser, nil, nil
		},
		UpdateUserRoleFunc: func(ctx context.Context, userUID string, roleID int) (*client.UpdateUserRoleResponse, annotations.Annotations, error) {
			return &client.UpdateUserRoleResponse{Message: "Role updated"}, nil, nil
		},
	}
	builder := &roleBuilderTest{
		resourceType: roleResourceType,
		client:       mockCli,
	}
	roleRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: roleResourceType.Id,
			Resource:     "ADMIN",
		},
		DisplayName: "Administrator",
	}
	userRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: userResourceType.Id,
			Resource:     "user-1",
		},
	}
	ent := &v2.Entitlement{
		Resource: roleRes,
	}
	grants, annos, err := builder.Grant(context.Background(), userRes, ent)
	assert.NoError(t, err)
	assert.NotNil(t, grants)
	assert.Equal(t, 0, len(annos))
}

// TestRoleBuilder_Grant_Error tests the Grant method of roleBuilderTest to ensure an error is returned if the user is not found.
func TestRoleBuilder_Grant_Error(t *testing.T) {
	mockCli := &test.MockClient{
		GetUserByIDFunc: func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
			return nil, nil, errors.New("mock error")
		},
	}
	builder := &roleBuilderTest{
		resourceType: roleResourceType,
		client:       mockCli,
	}
	roleRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: roleResourceType.Id,
			Resource:     "ADMIN",
		},
		DisplayName: "Administrator",
	}
	userRes := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: userResourceType.Id,
			Resource:     "user-1",
		},
	}
	ent := &v2.Entitlement{
		Resource: roleRes,
	}
	grants, annos, err := builder.Grant(context.Background(), userRes, ent)
	assert.Error(t, err)
	assert.Nil(t, grants)
	assert.Nil(t, annos)
}

// TestRoleBuilder_Revoke tests the Revoke method of roleBuilderTest to ensure a role is removed from a user.
func TestRoleBuilder_Revoke(t *testing.T) {
	mockUser := &client.ZuperUser{
		UserUID:   "user-1",
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		Role:      &client.Role{RoleUID: "2", RoleKey: "TEAM_LEADER", RoleName: "Team Leader"},
	}
	mockCli := &test.MockClient{
		GetUserByIDFunc: func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
			return mockUser, nil, nil
		},
		UpdateUserRoleFunc: func(ctx context.Context, userUID string, roleID int) (*client.UpdateUserRoleResponse, annotations.Annotations, error) {
			return &client.UpdateUserRoleResponse{Message: "Role updated to default"}, nil, nil
		},
	}
	builder := &roleBuilderTest{
		resourceType: roleResourceType,
		client:       mockCli,
	}
	grant := &v2.Grant{
		Principal:   &v2.Resource{Id: &v2.ResourceId{ResourceType: userResourceType.Id, Resource: "user-1"}},
		Entitlement: &v2.Entitlement{Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: roleResourceType.Id, Resource: "TEAM_LEADER"}}},
	}
	annos, err := builder.Revoke(context.Background(), grant)
	assert.NoError(t, err)
	assert.Nil(t, annos)
}

// TestRoleBuilder_Revoke_Error tests the Revoke method of roleBuilderTest to ensure an error is returned if the user is not found.
func TestRoleBuilder_Revoke_Error(t *testing.T) {
	mockCli := &test.MockClient{
		GetUserByIDFunc: func(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error) {
			return nil, nil, errors.New("mock error")
		},
	}
	builder := &roleBuilderTest{
		resourceType: roleResourceType,
		client:       mockCli,
	}
	grant := &v2.Grant{
		Principal:   &v2.Resource{Id: &v2.ResourceId{ResourceType: userResourceType.Id, Resource: "user-1"}},
		Entitlement: &v2.Entitlement{Resource: &v2.Resource{Id: &v2.ResourceId{ResourceType: roleResourceType.Id, Resource: "TEAM_LEADER"}}},
	}
	annos, err := builder.Revoke(context.Background(), grant)
	assert.Error(t, err)
	assert.Nil(t, annos)
}
