package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/stretchr/testify/assert"
)

// Mock response that simulates the expected structure of UsersResponse.
var mockUsersResponse = UsersResponse{
	CurrentPage: 1,
	TotalPages:  2,
	Data: []ZuperUser{
		{
			UserUID:   "123",
			FirstName: "Juan",
			LastName:  "Pérez",
			Email:     "juan@example.com",
			IsActive:  true,
		},
		{
			UserUID:   "456",
			FirstName: "Ana",
			LastName:  "García",
			Email:     "ana@example.com",
			IsActive:  true,
		},
	},
}

// TestGetUsers verifies that the GetUsers method correctly fetches
// and parses the list of users from the API, handles pagination tokens,
// and returns expected annotations without errors.
func TestGetUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/user/all?limit=50&page=1", r.URL.String())

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "99")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockUsersResponse)
	}))
	defer server.Close()

	ctx := context.Background()
	httpClient, err := uhttp.NewBaseHttpClientWithContext(ctx, &http.Client{})
	assert.NoError(t, err)

	client := NewClient(ctx, server.URL, "dummy-token", httpClient)

	page := EncodePageToken(&pageToken{Page: 1})

	users, nextPageToken, annos, err := client.GetUsers(ctx, page)
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	expectedNextToken := EncodePageToken(&pageToken{Page: 2})
	assert.Equal(t, expectedNextToken, nextPageToken)

	assert.IsType(t, annotations.Annotations{}, annos)

	assert.Equal(t, "123", users[0].UserUID)
	assert.Equal(t, "Juan", users[0].FirstName)
	assert.Equal(t, "Pérez", users[0].LastName)
	assert.Equal(t, "juan@example.com", users[0].Email)
}

// TestDoRequestInvalidURL tests the behavior of the doRequest method
// when provided with an invalid URL, expecting it to return an error.
func TestDoRequestInvalidURL(t *testing.T) {
	ctx := context.Background()
	httpClient, err := uhttp.NewBaseHttpClientWithContext(ctx, &http.Client{})
	assert.NoError(t, err)

	client := NewClient(ctx, "http://invalid-url", "token", httpClient)

	_, _, err = client.doRequest(ctx, http.MethodGet, "::bad_url::", nil)
	assert.Error(t, err)
}
