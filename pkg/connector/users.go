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

type UserClient interface {
	GetUsers(ctx context.Context, token string) ([]client.ZuperUser, string, annotations.Annotations, error)
}

// ResourceType returns the resource type for the user builder.
type userBuilder struct {
	resourceType *v2.ResourceType
	client       UserClient
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List retrieves all users and converts them into Baton resources.
// It handles pagination using the provided token and collects any annotations.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resource []*v2.Resource
	annos := annotations.Annotations{}

	users, nextToken, newAnnos, err := o.client.GetUsers(ctx, pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}

	for _, a := range newAnnos {
		annos.Append(a)
	}

	for _, user := range users {
		userCopy := user
		userResource, err := parseIntoUserResource(&userCopy, nil)
		if err != nil {
			return nil, "", nil, err
		}
		resource = append(resource, userResource)
	}

	return resource, nextToken, annos, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// parseIntoUserResource converts a ZuperUser into a Baton resource with user traits.
func parseIntoUserResource(user *client.ZuperUser, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	var userStatus v2.UserTrait_Status_Status
	if user.IsDeleted || !user.IsActive {
		userStatus = v2.UserTrait_Status_STATUS_DELETED
	} else {
		userStatus = v2.UserTrait_Status_STATUS_ENABLED
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
		"Role":        user.Role.RoleName,
	}

	if user.AccessRole != nil {
		profile["AccessRoleName"] = user.AccessRole.AccessRoleName
	}

	userTraits := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithStatus(userStatus),
		resource.WithUserLogin(user.Email),
	}

	displayName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	ret, err := resource.NewUserResource(
		displayName,
		userResourceType,
		user.UserUID,
		userTraits,
		resource.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// newUserBuilder creates a new instance of the userBuilder.
func newUserBuilder(client *client.Client) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}
