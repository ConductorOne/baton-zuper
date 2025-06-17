package connector

import (
	"context"
	"os"
	"testing"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/conductorone/baton-zuper/pkg/client"
	"github.com/stretchr/testify/assert"
)

var (
	pageToken = ""
)

// initClient initializes a Zuper client using environment variables for integration tests.
// Skips the tests if any required environment variable is missing.
func initClient(t *testing.T) *client.Client {
	ctx := context.Background()

	apiURL, apiOK := os.LookupEnv("ZUPER_API_URL")
	apiKey, keyOK := os.LookupEnv("ZUPER_API_KEY")

	if !apiOK || !keyOK {
		t.Skip("Missing ZUPER_API_URL or ZUPER_API_KEY environment variable. Skipping integration test.")
	}

	httpClient, err := uhttp.NewBaseHttpClientWithContext(ctx, nil)
	assert.NoError(t, err)

	return client.NewClient(ctx, apiURL, apiKey, httpClient)
}

// TestGetUsers verifies that users can be listed successfully from the Zuper API.
func TestGetUsers(t *testing.T) {
	ctx := context.Background()
	c := initClient(t)

	users, nextPage, _, err := c.GetUsers(ctx, client.PageOptions{
		PageSize:  client.DefaultPageSize,
		PageToken: pageToken,
	})

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

// TestGetTeams verifies that teams can be listed successfully from the Zuper API.
func TestGetTeams(t *testing.T) {
	ctx := context.Background()
	c := initClient(t)

	teams, nextPage, _, err := c.GetTeams(ctx, client.PageOptions{
		PageSize:  client.DefaultPageSize,
		PageToken: pageToken,
	})

	assert.NoError(t, err)
	assert.NotNil(t, teams)
	t.Logf("Retrieved %d teams, next page token: %s", len(teams), nextPage)
}

// TestGetTeamUsers verifies that team users can be listed successfully from the Zuper API.
func TestGetTeamUsers(t *testing.T) {
	ctx := context.Background()
	c := initClient(t)

	// You may want to get a real team ID from the API or use a known one for your test environment.
	teams, _, _, err := c.GetTeams(ctx, client.PageOptions{PageSize: 1})
	if err != nil || len(teams) == 0 {
		t.Skip("No teams available to test team users.")
	}
	teamID := teams[0].TeamUID

	users, _, _, err := c.GetTeamUsers(ctx, teamID)
	assert.NoError(t, err)
	assert.NotNil(t, users)
	t.Logf("Team %s has %d users", teamID, len(users))
}
