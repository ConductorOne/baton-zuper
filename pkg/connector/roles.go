package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-zuper/pkg/client"
)

type roleDefinition struct {
	ID          string
	DisplayName string
	Description string
	RoleKey     string
}

const (
	roleResourceID = "zuper-roles"
)

// roleDefinition{ role_id, role_name, role_descripcion, role_key}.
var roleDefinitions = []roleDefinition{
	{"1", "Administrator", "Indicates some actions are exclusive for admins", "ADMIN"},
	{"2", "Team Leader", "Indicates some actions are exclusive for team leaders", "TEAM_LEADER"},
	{"3", "Field Executive", "Indicates some actions are exclusive for field executives", "FIELD_EXECUTIVE"},
}

// roleBuilder manages role resources and their entitlements.
type roleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

// ResourceType returns the resource type managed by this builder.
func (r *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return r.resourceType
}

// List returns a singleton resource for all defined roles, simulating pagination.
func (r *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	annos := annotations.Annotations{}
	bag, pageToken, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: roleResourceType.Id})
	if err != nil {
		return nil, "", annos, err
	}
	roleResource, err := resource.NewRoleResource(
		roleResourceID,
		r.resourceType,
		roleResourceID,
		nil,
	)
	if err != nil {
		return nil, "", annos, fmt.Errorf("failed to create role resource: %w", err)
	}

	var outToken string
	if pageToken == "" {
		outToken, err = bag.NextToken("end")
		if err != nil {
			return nil, "", annos, err
		}
		return []*v2.Resource{roleResource}, outToken, annos, nil
	}
	return []*v2.Resource{}, "", annos, nil
}

// GetRoleResource returns the singleton role resource.
func (r *roleBuilder) GetRoleResource(ctx context.Context) (*v2.Resource, error) {
	roleResource, err := resource.NewRoleResource(
		roleResourceID,
		r.resourceType,
		roleResourceID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create role resource: %w", err)
	}
	return roleResource, nil
}

// Entitlements returns all entitlements for the given role resource.
func (r *roleBuilder) Entitlements(ctx context.Context, roleRes *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	annos := annotations.Annotations{}
	entitlements := make([]*v2.Entitlement, 0, len(roleDefinitions))
	for _, role := range roleDefinitions {
		entitlements = append(entitlements, entitlement.NewPermissionEntitlement(
			roleRes,
			role.RoleKey,
			entitlement.WithDisplayName(role.DisplayName),
			entitlement.WithDescription(role.Description),
			entitlement.WithGrantableTo(userResourceType),
		))
	}

	return entitlements, "", annos, nil
}

// Grants returns the grants for a role resource (none in this implementation).
func (r *roleBuilder) Grants(ctx context.Context, roleRes *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// newRoleBuilder creates a new instance of roleBuilder.
func newRoleBuilder(client *client.Client) *roleBuilder {
	return &roleBuilder{
		resourceType: roleResourceType,
		client:       client,
	}
}
