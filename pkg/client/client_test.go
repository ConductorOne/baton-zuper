package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/stretchr/testify/assert"
)

// loadUsersResponseFromMock loads a UsersResponse from a mock JSON file for testing.
func loadUsersResponseFromMock(file string) UsersResponse {
	var users []ZuperUser
	mockData, err := os.ReadFile("../../test/mock/" + file)
	if err != nil {
		panic(err)
	}
	_ = json.Unmarshal(mockData, &users)
	return UsersResponse{
		CurrentPage: 1,
		TotalPages:  1,
		Data:        users,
	}
}

func TestGetUsers(t *testing.T) {
	t.Run("success, single page", func(t *testing.T) {
		mockResp := loadUsersResponseFromMock("users_success.json")
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.String(), "/api/user/all")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(mockResp)
		}))
		defer server.Close()

		ctx := context.Background()
		httpClient, _ := uhttp.NewBaseHttpClientWithContext(ctx, &http.Client{})
		client := NewClient(ctx, server.URL, "dummy-token", httpClient)

		users, nextPageToken, annos, err := client.GetUsers(ctx, PageOptions{
			PageSize:  DefaultPageSize,
			PageToken: "",
		})
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Empty(t, nextPageToken)
		assert.IsType(t, annotations.Annotations{}, annos)
	})

	t.Run("success, paginated", func(t *testing.T) {
		mockResp1 := loadUsersResponseFromMock("users_success.json")
		mockResp1.TotalPages = 2
		mockResp2 := loadUsersResponseFromMock("users_success.json")
		mockResp2.CurrentPage = 2
		mockResp2.TotalPages = 2
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if calls == 1 {
				_ = json.NewEncoder(w).Encode(mockResp1)
			} else {
				_ = json.NewEncoder(w).Encode(mockResp2)
			}
		}))
		defer server.Close()

		ctx := context.Background()
		httpClient, _ := uhttp.NewBaseHttpClientWithContext(ctx, &http.Client{})
		client := NewClient(ctx, server.URL, "dummy-token", httpClient)

		// First page
		users, nextPageToken, _, err := client.GetUsers(ctx, PageOptions{
			PageSize:  DefaultPageSize,
			PageToken: "",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, nextPageToken)
		assert.Len(t, users, 1)
		// Second page
		users2, nextPageToken2, _, err := client.GetUsers(ctx, PageOptions{
			PageSize:  DefaultPageSize,
			PageToken: nextPageToken,
		})
		assert.NoError(t, err)
		assert.Empty(t, nextPageToken2)
		assert.Len(t, users2, 1)
	})

	t.Run("error, invalid URL", func(t *testing.T) {
		ctx := context.Background()
		httpClient, _ := uhttp.NewBaseHttpClientWithContext(ctx, &http.Client{})
		client := NewClient(ctx, "::bad_url::", "token", httpClient)
		_, _, err := client.doRequest(ctx, http.MethodGet, "::bad_url::", nil)
		assert.Error(t, err)
	})

	t.Run("error, server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		ctx := context.Background()
		httpClient, _ := uhttp.NewBaseHttpClientWithContext(ctx, &http.Client{})
		client := NewClient(ctx, server.URL, "dummy-token", httpClient)
		_, _, _, err := client.GetUsers(ctx, PageOptions{PageSize: DefaultPageSize})
		assert.Error(t, err)
	})
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

func TestGetUserByID(t *testing.T) {
	t.Run("success, user details", func(t *testing.T) {
		mockData, err := os.ReadFile("../../test/mock/user_details_success.json")
		assert.NoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.String(), "/api/user/")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(mockData)
		}))
		defer server.Close()

		ctx := context.Background()
		httpClient, _ := uhttp.NewBaseHttpClientWithContext(ctx, &http.Client{})
		client := NewClient(ctx, server.URL, "dummy-token", httpClient)

		user, annos, err := client.GetUserByID(ctx, "c3dea3e3-8bc3-459f-aaeb-04fd6f501fa5")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Ramon", user.FirstName)
		assert.Equal(t, "Mendoza", user.LastName)
		assert.Equal(t, "Ramon.Mendoza@Powin.com", user.Email)
		assert.IsType(t, annotations.Annotations{}, annos)
	})

	t.Run("error, not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		ctx := context.Background()
		httpClient, _ := uhttp.NewBaseHttpClientWithContext(ctx, &http.Client{})
		client := NewClient(ctx, server.URL, "dummy-token", httpClient)

		user, annos, err := client.GetUserByID(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Nil(t, annos)
	})
}
