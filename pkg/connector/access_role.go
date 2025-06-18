package connector

import (
	"context"
	"fmt"
	"sync"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-zuper/pkg/client"
)

const (
	cacheTTL = 5 * time.Minute
)

// accessRoleBuilder manages access role resources and their entitlements.
type accessRoleBuilder struct {
	resourceType *v2.ResourceType
	client       UserClient
	mu           sync.RWMutex
	roleCache    map[string]*client.AccessRole
	lastFetch    time.Time
}

// newAccessRoleBuilder creates a new accessRoleBuilder instance.
func newAccessRoleBuilder(client UserClient) *accessRoleBuilder {
	return &accessRoleBuilder{
		resourceType: accessRoleResourceType,
		client:       client,
	}
}

// UpdateCacheWithUsers updates the access role cache with roles from a list of users.
func (b *accessRoleBuilder) updateCacheWithUsers(users []*client.ZuperUser) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.roleCache == nil {
		b.roleCache = make(map[string]*client.AccessRole)
	}
	for _, user := range users {
		if user.AccessRole != nil {
			b.roleCache[user.AccessRole.AccessRoleUID] = user.AccessRole
		}
	}
	b.lastFetch = time.Now()
}

// loadAccessRoles loads all access roles from all users, handling pagination.
func (b *accessRoleBuilder) loadAccessRoles(ctx context.Context) error {
	b.mu.RLock()
	if time.Since(b.lastFetch) < cacheTTL && len(b.roleCache) > 0 {
		b.mu.RUnlock()
		return nil
	}
	b.mu.RUnlock()

	var allUsers []*client.ZuperUser
	pToken := ""
	for {
		users, nextToken, _, err := b.client.GetUsers(ctx, client.PageOptions{
			PageToken: pToken,
			PageSize:  client.DefaultPageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to load users for access role cache: %w", err)
		}
		allUsers = append(allUsers, users...)
		if nextToken == "" {
			break
		}
		pToken = nextToken
	}

	b.updateCacheWithUsers(allUsers)
	return nil
}

// ResourceType returns the resource type for access roles.
func (b *accessRoleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return b.resourceType
}

// List returns all access roles as individual resources, simulating pagination.
func (b *accessRoleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	annos := annotations.Annotations{}

	err := b.loadAccessRoles(ctx)
	if err != nil {
		return nil, "", annos, err
	}

	var resources []*v2.Resource
	for uid, role := range b.roleCache {
		profile := map[string]interface{}{
			"AccessRoleUID":   role.AccessRoleUID,
			"RoleDescription": role.RoleDescription,
		}
		accessRoleResource, err := resource.NewRoleResource(
			role.AccessRoleName,
			b.resourceType,
			uid,
			[]resource.RoleTraitOption{resource.WithRoleProfile(profile)},
		)
		if err != nil {
			return nil, "", annos, fmt.Errorf("failed to create access role resource: %w", err)
		}
		resources = append(resources, accessRoleResource)
	}

	return resources, "", annos, nil
}

// Entitlements returns an 'assigned' entitlement for the given access role resource.
func (b *accessRoleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	annos := annotations.Annotations{}
	name := resource.DisplayName
	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(fmt.Sprintf("%s to %s access role", assignedEntitlement, name)),
		entitlement.WithDisplayName(fmt.Sprintf("%s access role %s", name, assignedEntitlement)),
	}
	entitlements := []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(resource, assignedEntitlement, assigmentOptions...),
	}
	return entitlements, "", annos, nil
}

// Grants would assign access roles to users. This is intentionally left empty as grants are handled by the userBuilder.
func (b *accessRoleBuilder) Grants(ctx context.Context, resourceUser *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grant assigns an access role to a user if the user does not already have it. Used for access role provisioning.
func (b *accessRoleBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
	userID := principal.Id.Resource
	accessRoleUID := entitlement.Resource.Id.Resource

	user, _, err := b.client.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user.AccessRole != nil && user.AccessRole.AccessRoleUID == accessRoleUID {
		return nil, annotations.New(&v2.GrantAlreadyExists{}), nil
	}

	resp, annos, err := b.client.UpdateUserAccessRole(ctx, userID, accessRoleUID)
	if err != nil {
		return nil, annos, fmt.Errorf("failed to update user access role: %w", err)
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

// Revoke removes an access role from a user by setting it to an empty string. Used for access role deprovisioning.
func (b *accessRoleBuilder) Revoke(ctx context.Context, g *v2.Grant) (annotations.Annotations, error) {
	userID := g.Principal.Id.Resource

	user, _, err := b.client.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user.AccessRole == nil {
		return annotations.New(&v2.GrantAlreadyRevoked{}), nil
	}

	_, annos, err := b.client.UpdateUserAccessRole(ctx, userID, "")
	if err != nil {
		return annos, fmt.Errorf("failed to remove user access role: %w", err)
	}
	return annos, nil
}
