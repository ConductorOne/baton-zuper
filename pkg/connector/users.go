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
	GetUsers(ctx context.Context, pToken string) ([]*client.ZuperUser, string, annotations.Annotations, error)
}

type userBuilder struct {
	resourceType *v2.ResourceType
	client       UserClient
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	users, nextToken, annos, err := o.client.GetUsers(ctx, pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}

	var resources []*v2.Resource
	for _, user := range users {
		userResource, err := parseIntoUserResource(user, nil)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, userResource)
	}

	return resources, nextToken, annos, nil
}

func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func parseIntoUserResource(user *client.ZuperUser, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	userStatus := v2.UserTrait_Status_STATUS_ENABLED
	if user.IsDeleted || !user.IsActive {
		userStatus = v2.UserTrait_Status_STATUS_DISABLED
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
		resource.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func newUserBuilder(client *client.Client) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}
