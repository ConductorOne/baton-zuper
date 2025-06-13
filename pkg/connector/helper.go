package connector

import (
	"errors"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/crypto"
	"github.com/conductorone/baton-sdk/pkg/pagination"
)

// assignedEntitlement is the standard entitlement name for assignments. symbols is the set of special characters allowed in the password.
const (
	assignedEntitlement = "assigned"
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

// generateCredentials generates a random password based on the credential options.
func generateCredentials(credentialOptions *v2.CredentialOptions) (string, error) {
	if credentialOptions == nil || credentialOptions.GetRandomPassword() == nil {
		return "", errors.New("unsupported credential option: only random password is supported")
	}

	length := credentialOptions.GetRandomPassword().GetLength()
	if length < 12 {
		length = 12
	}

	password, err := crypto.GenerateRandomPassword(
		&v2.CredentialOptions_RandomPassword{
			Length: length,
		},
	)
	if err != nil {
		return "", err
	}
	return password, nil
}
