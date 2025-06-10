package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-zuper/pkg/client"
)

// UserClient defines the interface for fetching users with pagination options.
type UserClient interface {
	GetUsers(ctx context.Context, options client.PageOptions) ([]*client.ZuperUser, string, annotations.Annotations, error)
}

type userBuilder struct {
	resourceType      *v2.ResourceType
	client            UserClient
	roleBuilder       *roleBuilder
	accessRoleBuilder *accessRoleBuilder
}

// ResourceType returns the resource type for users.
func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns a paginated list of user resources.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource
	bag, pageToken, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: userResourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}
	users, nextPageToken, annotation, err := o.client.GetUsers(ctx, client.PageOptions{
		PageSize:  pToken.Size,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", nil, err
	}

	for _, user := range users {
		userResource, err := parseIntoUserResource(user)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, userResource)
	}
	var outToken string
	if nextPageToken != "" {
		outToken, err = bag.NextToken(nextPageToken)
		if err != nil {
			return nil, "", nil, err
		}
	}

	return resources, outToken, annotation, nil
}

// Entitlements returns the entitlements for a user resource (none in this implementation).
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants returns the grants for a user resource, including roles and access roles.
func (o *userBuilder) Grants(ctx context.Context, resourceUser *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	// Grants Role.
	roleKey, annos, err := o.findUserRoleKey(ctx, resourceUser.Id.Resource, pToken)
	if err != nil {
		return nil, "", annos, err
	}

	roleResource, err := o.roleBuilder.GetRoleResource(ctx)
	if err != nil {
		return nil, "", annos, fmt.Errorf("failed to get role resource: %w", err)
	}

	grants, err := createUserGrants(roleResource, resourceUser, roleKey)
	if err != nil {
		return nil, "", annos, fmt.Errorf("failed to create grants for %s: %w", resourceUser.Id.Resource, err)
	}

	// Grants AccessRole.
	accessRoleUID, annos2, err := o.findUserAccessRoleUID(ctx, resourceUser.Id.Resource, pToken)
	if err == nil && accessRoleUID != "" && o.accessRoleBuilder != nil {
		accessRoleResource, err := resource.NewRoleResource(
			accessRoleEntitlementPrefix,
			o.accessRoleBuilder.ResourceType(ctx),
			accessRoleEntitlementPrefix,
			nil,
		)
		if err == nil {
			accessRoleGrants, _ := createUserAccessRoleGrants(accessRoleResource, resourceUser, accessRoleUID)
			grants = append(grants, accessRoleGrants...)
		}
		for _, annon := range annos2 {
			annos.Append(annon)
		}
	}

	return grants, "", annos, nil
}

// parseIntoUserResource converts a ZuperUser into a v2.Resource for Baton.
func parseIntoUserResource(user *client.ZuperUser) (*v2.Resource, error) {
	userStatus := v2.UserTrait_Status_STATUS_ENABLED
	if !user.IsActive {
		userStatus = v2.UserTrait_Status_STATUS_DISABLED
	}
	if user.IsDeleted {
		userStatus = v2.UserTrait_Status_STATUS_DELETED
	}

	profile := map[string]interface{}{
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
		"Designation": user.Designation,
		"IsActive":    user.IsActive,
		"IsDeleted":   user.IsDeleted,
		"EmpCode":     user.EmpCode,
		"CreatedAt":   user.CreatedAt,
		"UpdatedAt":   user.UpdatedAt,
	}

	userTraits := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithStatus(userStatus),
		resource.WithUserLogin(user.Email),
		resource.WithEmail(user.Email, true),
	}

	displayName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	ret, err := resource.NewUserResource(
		displayName,
		userResourceType,
		user.UserUID,
		userTraits,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// newUserBuilder creates a new userBuilder instance.
func newUserBuilder(client UserClient, roleBuilder *roleBuilder, accessRoleBuilder *accessRoleBuilder) *userBuilder {
	return &userBuilder{
		resourceType:      userResourceType,
		client:            client,
		roleBuilder:       roleBuilder,
		accessRoleBuilder: accessRoleBuilder,
	}
}
