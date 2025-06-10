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
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-zuper/pkg/client"
)

const (
	accessRoleEntitlementPrefix = "access-role"
	cacheTTL                    = 5 * time.Minute
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
func newAccessRoleBuilder(client UserClient, roleCache map[string]*client.AccessRole) *accessRoleBuilder {
	return &accessRoleBuilder{
		resourceType: accessRoleResourceType,
		client:       client,
		roleCache:    roleCache,
	}
}

// UpdateCacheWithUsers updates the access role cache with roles from a list of users.
func (b *accessRoleBuilder) UpdateCacheWithUsers(users []*client.ZuperUser) {
	b.mu.Lock()
	defer b.mu.Unlock()

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

	b.UpdateCacheWithUsers(allUsers)
	return nil
}

// GetAccessRole retrieves an access role from the cache by UID.
func (b *accessRoleBuilder) GetAccessRole(ctx context.Context, uid string) (*client.AccessRole, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	role, ok := b.roleCache[uid]
	if !ok {
		return nil, fmt.Errorf("access role not found: %s", uid)
	}
	return role, nil
}

// ResourceType returns the resource type for access roles.
func (b *accessRoleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return b.resourceType
}

// List returns the singleton access role resource, simulating pagination.
func (b *accessRoleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	annos := annotations.Annotations{}
	bag, pageToken, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: accessRoleResourceType.Id})
	if err != nil {
		return nil, "", annos, err
	}
	accessRoleResource, err := resource.NewRoleResource(
		accessRoleEntitlementPrefix,
		b.resourceType,
		accessRoleEntitlementPrefix,
		nil,
	)
	if err != nil {
		return nil, "", annos, fmt.Errorf("failed to create access role resource: %w", err)
	}

	var outToken string
	if pageToken == "" {
		outToken, err = bag.NextToken("end")
		if err != nil {
			return nil, "", annos, err
		}
		return []*v2.Resource{accessRoleResource}, outToken, annos, nil
	}
	return []*v2.Resource{}, "", annos, nil
}

// Entitlements returns all access role entitlements from the cache.
func (b *accessRoleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	err := b.loadAccessRoles(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	var entitlements []*v2.Entitlement
	for _, role := range b.roleCache {
		ent := entitlement.NewPermissionEntitlement(
			resource,
			role.AccessRoleUID,
			entitlement.WithDisplayName(role.AccessRoleName),
			entitlement.WithDescription(role.RoleDescription),
			entitlement.WithGrantableTo(userResourceType),
		)
		entitlements = append(entitlements, ent)
	}

	return entitlements, "", nil, nil
}

// Grants returns the grants for an access role resource (none in this implementation).
func (b *accessRoleBuilder) Grants(ctx context.Context, resourceUser *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
