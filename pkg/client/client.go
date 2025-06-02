package client

import (
	"context"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/ratelimit"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

// API endpoint constant.
const (
	getUsers = "/api/user/all"
)

// Client struct stores configuration and HTTP client wrapper for API requests.
type Client struct {
	apiUrl  string
	Token   string
	wrapper *uhttp.BaseHttpClient
}

// New creates and returns a new Client using an existing Client's configuration.
// It initializes a uhttp client with logging enabled and wraps it with a BaseHttpClient.
func New(ctx context.Context, client *Client) (*Client, error) {
	var (
		clientApi   = client.apiUrl
		clientToken = client.Token
	)

	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	cli, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		wrapper: cli,
		apiUrl:  clientApi,
		Token:   clientToken,
	}, nil
}

// NewClient creates a new Client with the provided API URL and token.
// Optionally accepts a custom BaseHttpClient (for testing or reuse).
func NewClient(ctx context.Context, apiUrl string, token string, httpClient ...*uhttp.BaseHttpClient) *Client {
	var wrapper = &uhttp.BaseHttpClient{}
	if len(httpClient) != 0 {
		wrapper = httpClient[0]
	}
	return &Client{
		wrapper: wrapper,
		apiUrl:  apiUrl,
		Token:   token,
	}
}

// GetUsers fetches a paginated list of users from the API.
// It returns a slice of users, a next page token (if available), rate limit annotations, and an error if any.
func (c *Client) GetUsers(ctx context.Context, token string) ([]ZuperUser, string, annotations.Annotations, error) {
	opts := pageOptions{
		PageToken: token,
		PageSize:  DefaultPageSize,
	}

	usersURL, _, err := PreparePagedRequest(c.apiUrl, getUsers, opts)
	if err != nil {
		return nil, "", nil, err
	}

	var usersResponse UsersResponse
	headers, _, err := c.doRequest(ctx, http.MethodGet, usersURL.String(), &usersResponse)
	if err != nil {
		return nil, "", nil, err
	}
	annos := annotations.Annotations{}
	if desc, err := ratelimit.ExtractRateLimitData(http.StatusOK, &headers); err == nil {
		annos.WithRateLimiting(desc)
	}

	nextToken := GetNextToken(usersResponse.CurrentPage, usersResponse.TotalPages)

	return usersResponse.Data, nextToken, annos, nil
}

// doRequest builds and sends an HTTP request with common headers and handles the response.
// Supports GET, POST, PUT, and DELETE methods. Optionally parses the JSON response into 'res'.
func (c *Client) doRequest(
	ctx context.Context,
	method string,
	requestURL string,
	res interface{},
) (http.Header, annotations.Annotations, error) {
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.wrapper.NewRequest(
		ctx,
		method,
		parsedURL,
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithHeader("x-api-key", c.Token),
	)
	if err != nil {
		return nil, nil, err
	}

	var resp *http.Response
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut:
		var doOptions []uhttp.DoOption
		if res != nil {
			doOptions = append(doOptions, uhttp.WithJSONResponse(res))
		}
		resp, err = c.wrapper.Do(req, doOptions...)
	case http.MethodDelete:
		resp, err = c.wrapper.Do(req)
	}

	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	annotation := annotations.Annotations{}
	if desc, err := ratelimit.ExtractRateLimitData(resp.StatusCode, &resp.Header); err == nil {
		annotation.WithRateLimiting(desc)
	}

	return resp.Header, annotation, nil
}
