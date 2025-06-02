package connector

import (
	"context"
	"os"
	"testing"

	"github.com/conductorone/baton-zuper/pkg/client"
	"github.com/stretchr/testify/assert"
)

var (
	pageToken = ""
)

// initClient initializes a Zuper client using environment variables.
// Skips the tests if any required environment variable is missing.
func initClient(t *testing.T) *client.Client {
	ctx := context.Background()

	apiURL, apiOK := os.LookupEnv("ZUPER_API_URL")
	apiKey, keyOK := os.LookupEnv("ZUPER_API_KEY")

	if !apiOK || !keyOK {
		t.Skip("Missing ZUPER_API_URL or ZUPER_API_KEY environment variable. Skipping integration test.")
	}

	return client.NewClient(ctx, apiURL, apiKey)
}

// TestGetUsers verifies that users can be listed successfully from the Zuper API.
func TestGetUsers(t *testing.T) {
	ctx := context.Background()
	c := initClient(t)

	users, nextPage, _, err := c.GetUsers(ctx, pageToken)

	assert.NoError(t, err)
	assert.NotNil(t, users)
	t.Logf("Retrieved %d users, next page token: %s", len(users), nextPage)
}

// TestUserBuilderList verifies that the List method of the userBuilder retrieves users without error.
// It checks that the returned list is not nil and logs the number of users and pagination token.
func TestUserBuilderList(t *testing.T) {
	ctx := context.Background()
	client := initClient(t)

	ub := newUserBuilder(client)
	users, nextToken, _, err := ub.List(ctx, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, users)

	t.Logf("Users retrieved: %d, next token: %v", len(users), nextToken)
}
