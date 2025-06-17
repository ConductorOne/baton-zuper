package connector

import (
	"context"
	"fmt"
	"strconv"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-zuper/pkg/client"
)

// roleDefinition holds static information about a role.
type roleDefinition struct {
	ID          string
	DisplayName string
	Description string
	RoleKey     string
}

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

// List returns all defined roles as individual resources, simulating pagination.
func (r *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	annos := annotations.Annotations{}

	var resources []*v2.Resource
	for _, role := range roleDefinitions {
		profile := map[string]interface{}{
			"role_id":          role.ID,
			"role_key":         role.RoleKey,
			"role_display":     role.DisplayName,
			"role_description": role.Description,
		}
		roleResource, err := resource.NewRoleResource(
			role.DisplayName,
			r.resourceType,
			role.RoleKey,
			[]resource.RoleTraitOption{resource.WithRoleProfile(profile)},
		)
		if err != nil {
			return nil, "", annos, fmt.Errorf("failed to create role resource: %w", err)
		}
		resources = append(resources, roleResource)
	}
	return resources, "", annos, nil
}

// Entitlements returns an 'assigned' entitlement for the given role resource.
func (r *roleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	annos := annotations.Annotations{}

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(fmt.Sprintf("%s to %s role", assignedEntitlement, resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("%s role %s", resource.DisplayName, assignedEntitlement)),
	}

	ent := entitlement.NewAssignmentEntitlement(
		resource,
		assignedEntitlement,
		assigmentOptions...,
	)

	return []*v2.Entitlement{ent}, "", annos, nil
}

// Grants would assign roles to users. This is intentionally left empty as grants are handled by the userBuilder.
func (r *roleBuilder) Grants(ctx context.Context, roleRes *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grant assigns a role to a user if the user does not already have it. Used for role provisioning.
func (r *roleBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
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
		return nil, nil, fmt.Errorf("role ID not found for key: %s", roleKey)
	}
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid role ID: %s", roleIDStr)
	}

	user, _, err := r.client.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user.Role != nil && user.Role.RoleUID == roleIDStr {
		return nil, nil, nil
	}

	resp, annos, err := r.client.UpdateUserRole(ctx, userID, roleID)
	if err != nil {
		return nil, annos, fmt.Errorf("failed to update user role: %w", err)
	}

	grantObj := grant.NewGrant(
		entitlement.Resource,
		assignedEntitlement,
		principal.Id,
		grant.WithGrantMetadata(map[string]interface{}{
			"message": resp.Message,
		}),
	)
	return []*v2.Grant{grantObj}, annos, nil
}

// Revoke removes a role from a user and assigns the default role (Field Executive) if not already set. Used for role deprovisioning.
func (r *roleBuilder) Revoke(ctx context.Context, g *v2.Grant) (annotations.Annotations, error) {
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
		return nil, fmt.Errorf("role ID not found for key: %s", roleKey)
	}

	user, _, err := r.client.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// If the user already has the default role (3), do nothing
	if user.Role.RoleKey == "FIELD_EXECUTIVE" {
		return nil, nil
	}

	defaultRoleID := 3 // Field Executive
	_, annos, err := r.client.UpdateUserRole(ctx, userID, defaultRoleID)
	if err != nil {
		return annos, fmt.Errorf("failed to set default role: %w", err)
	}
	return annos, nil
}

// newRoleBuilder creates a new instance of roleBuilder.
func newRoleBuilder(client *client.Client) *roleBuilder {
	return &roleBuilder{
		resourceType: roleResourceType,
		client:       client,
	}
}
