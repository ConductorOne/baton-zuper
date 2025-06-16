package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	grantpkg "github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-zuper/pkg/client"
)

// UserClient defines the interface for fetching users with pagination options.
type UserClient interface {
	GetUsers(ctx context.Context, options client.PageOptions) ([]*client.ZuperUser, string, annotations.Annotations, error)
	GetUserByID(ctx context.Context, userUID string) (*client.ZuperUser, annotations.Annotations, error)
	CreateUser(ctx context.Context, user client.UserPayload) (*client.CreateUserResponse, annotations.Annotations, error)
}

type userBuilder struct {
	resourceType *v2.ResourceType
	client       UserClient
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
	annos := annotations.Annotations{}
	var grants []*v2.Grant

	user, _, err := o.client.GetUserByID(ctx, resourceUser.Id.Resource)
	if err != nil {
		return nil, "", annos, err
	}
	if user == nil {
		return nil, "", annos, nil
	}

	// Grant Role.
	if user.Role != nil {
		roleRes := makeRoleSubjectID(user.Role.RoleKey, user)
		userId := makeUserSubjectID(user.UserUID)
		grant := grantpkg.NewGrant(roleRes, assignedEntitlement, userId)
		grants = append(grants, grant)
	}

	// Grant AccessRole.
	if user.AccessRole != nil {
		accessRoleRes := makeAccessRoleSubjectID(user.AccessRole.AccessRoleUID, user)
		userId := makeUserSubjectID(user.UserUID)
		grant := grantpkg.NewGrant(accessRoleRes, assignedEntitlement, userId)
		grants = append(grants, grant)
	}

	return grants, "", annos, nil
}

// CreateAccountCapabilityDetails declares support for account provisioning with password.
func (u *userBuilder) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
	}, nil, nil
}

// CreateAccount provisions a new Zuper user based on AccountInfo and CredentialOptions.
func (u *userBuilder) CreateAccount(
	ctx context.Context,
	accountInfo *v2.AccountInfo,
	credentialOptions *v2.CredentialOptions,
) (
	connectorbuilder.CreateAccountResponse,
	[]*v2.PlaintextData,
	annotations.Annotations,
	error,
) {
	profile := accountInfo.GetProfile().AsMap()
	requiredFields := map[string]string{
		"first_name": "first_name is required",
		"last_name":  "last_name is required",
		"email":      "email is required",
		"emp_code":   "emp_code is required",
	}
	for field, errMsg := range requiredFields {
		if val, ok := profile[field].(string); !ok || val == "" {
			return nil, nil, nil, fmt.Errorf("%s", errMsg)
		}
	}

	generatedPassword, err := generateCredentials(credentialOptions)
	if err != nil {
		return nil, nil, nil, err
	}
	if generatedPassword == "" {
		return nil, nil, nil, fmt.Errorf("password is required for user creation")
	}
	userPayload := client.UserPayload{
		FirstName:   profile["first_name"].(string),
		LastName:    profile["last_name"].(string),
		Email:       profile["email"].(string),
		Password:    generatedPassword,
		Designation: "Field Executive",
		EmpCode:     profile["emp_code"].(string),
		RoleID:      "3",
	}

	resp, annos, err := u.client.CreateUser(ctx, userPayload)
	if err != nil {
		return nil, nil, annos, fmt.Errorf("failed to create user: %w", err)
	}

	newUser := &client.ZuperUser{
		UserUID:     resp.Data.UserUID,
		FirstName:   userPayload.FirstName,
		LastName:    userPayload.LastName,
		Email:       userPayload.Email,
		Designation: userPayload.Designation,
		EmpCode:     userPayload.EmpCode,
		IsActive:    true,
		IsDeleted:   false,
	}

	userResource, err := parseIntoUserResource(newUser)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse created user: %w", err)
	}

	plaintexts := []*v2.PlaintextData{{Bytes: []byte(generatedPassword)}}

	return &v2.CreateAccountResponse_SuccessResult{
		Resource: userResource,
	}, plaintexts, annos, nil
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

// makeUserSubjectID creates a ResourceId for a user based on their user ID.
func makeUserSubjectID(userID string) *v2.ResourceId {
	return &v2.ResourceId{
		ResourceType: userResourceType.Id,
		Resource:     userID,
	}
}

// makeRoleSubjectID creates a Resource for a role grant.
func makeRoleSubjectID(roleID string, user *client.ZuperUser) *v2.Resource {
	return &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: roleResourceType.Id,
			Resource:     roleID,
		},
		DisplayName: user.Role.RoleName,
	}
}

// makeAccessRoleSubjectID creates a Resource for an access role grant.
func makeAccessRoleSubjectID(accessRoleUID string, user *client.ZuperUser) *v2.Resource {
	return &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: accessRoleResourceType.Id,
			Resource:     accessRoleUID,
		},
		DisplayName: user.AccessRole.AccessRoleName,
	}
}

// newUserBuilder creates a new userBuilder instance.
func newUserBuilder(client UserClient) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}