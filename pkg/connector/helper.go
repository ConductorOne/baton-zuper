package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	grantpkg "github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-zuper/pkg/client"
)

// parsePageToken deserializes the Baton token and returns the Bag and page number for upstream.
func parsePageToken(i string, resourceID *v2.ResourceId) (*pagination.Bag, string, error) {
	b := &pagination.Bag{}
	if err := b.Unmarshal(i); err != nil {
		return nil, "", err
	}

	if b.Current() == nil {
		b.Push(pagination.PageState{
			ResourceTypeID: resourceID.ResourceType,
			ResourceID:     resourceID.Resource,
		})
	}

	return b, b.PageToken(), nil
}

// findUserRoleKey finds the role key for a user by userId, handling pagination.
func (o *userBuilder) findUserRoleKey(ctx context.Context, userId string, pToken *pagination.Token) (string, annotations.Annotations, error) {
	annos := annotations.Annotations{}
	token := pToken.Token

	for {
		users, nextToken, pageAnnos, err := o.client.GetUsers(ctx, client.PageOptions{PageToken: token, PageSize: client.DefaultPageSize})
		if err != nil {
			return "", annos, fmt.Errorf("failed to fetch users: %w", err)
		}

		for _, annon := range pageAnnos {
			annos.Append(annon)
		}

		for _, user := range users {
			if user.UserUID == userId {
				if user.Role.RoleKey == "" {
					return "", annos, fmt.Errorf("user %s has no role", userId)
				}
				return user.Role.RoleKey, annos, nil
			}
		}

		if nextToken == "" {
			break
		}
		token = nextToken
	}

	return "", annos, fmt.Errorf("user %s not found", userId)
}

// createUserGrants creates the grants for a user based on their role_key.
func createUserGrants(roleResource *v2.Resource, userResource *v2.Resource, roleKey string) ([]*v2.Grant, error) {
	var grants []*v2.Grant
	var matchedRole *roleDefinition
	for _, roleDefinition := range roleDefinitions {
		if roleDefinition.RoleKey == roleKey {
			matchedRole = &roleDefinition
			break
		}
	}
	if matchedRole == nil {
		return nil, nil
	}

	grantObj := grantpkg.NewGrant(
		roleResource,
		matchedRole.RoleKey,
		userResource,
	)
	grants = append(grants, grantObj)
	return grants, nil
}

// findUserAccessRoleUID finds the access role UID for a user by userId, handling pagination.
func (o *userBuilder) findUserAccessRoleUID(ctx context.Context, userId string, pToken *pagination.Token) (string, annotations.Annotations, error) {
	annos := annotations.Annotations{}
	token := pToken.Token

	for {
		users, nextToken, pageAnnos, err := o.client.GetUsers(ctx, client.PageOptions{PageToken: token, PageSize: client.DefaultPageSize})
		if err != nil {
			return "", annos, fmt.Errorf("failed to fetch users: %w", err)
		}

		for _, annon := range pageAnnos {
			annos.Append(annon)
		}

		for _, user := range users {
			if user.UserUID == userId {
				if user.AccessRole == nil || user.AccessRole.AccessRoleUID == "" {
					return "", annos, fmt.Errorf("user %s has no access role", userId)
				}
				return user.AccessRole.AccessRoleUID, annos, nil
			}
		}

		if nextToken == "" {
			break
		}
		token = nextToken
	}

	return "", annos, fmt.Errorf("user %s not found", userId)
}

// createUserAccessRoleGrants creates the grants for a user based on their access_role_uid.
func createUserAccessRoleGrants(accessRoleResource *v2.Resource, userResource *v2.Resource, accessRoleUID string) ([]*v2.Grant, error) {
	if accessRoleUID == "" {
		return nil, nil
	}
	grantObj := grantpkg.NewGrant(
		accessRoleResource,
		accessRoleUID,
		userResource,
	)
	return []*v2.Grant{grantObj}, nil
}
