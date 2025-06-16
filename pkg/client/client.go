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

const (
	userEndpoint = "/api/user"
)

// Client is the Zuper API client for Baton.
type Client struct {
	apiUrl  string
	apiKey  string
	wrapper *uhttp.BaseHttpClient
}

func New(ctx context.Context, client *Client) (*Client, error) {
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
		apiUrl:  client.apiUrl,
		apiKey:  client.apiKey,
	}, nil
}

// NewClient creates a new Client instance with the provided HTTP client.
func NewClient(ctx context.Context, apiUrl string, apiKey string, httpClient *uhttp.BaseHttpClient) *Client {
	if httpClient == nil {
		httpClient = &uhttp.BaseHttpClient{}
	}
	return &Client{
		wrapper: httpClient,
		apiUrl:  apiUrl,
		apiKey:  apiKey,
	}
}

// GetUsers fetches a paginated list of users from the Zuper API.
func (c *Client) GetUsers(ctx context.Context, opts PageOptions) ([]*ZuperUser, string, annotations.Annotations, error) {
	if opts.PageSize == 0 {
		opts.PageSize = DefaultPageSize
	}

	usersURL, _, err := preparePagedRequest(c.apiUrl, userEndpoint, opts, "all")
	if err != nil {
		return nil, "", nil, err
	}

	var usersResponse UsersResponse
	_, annos, err := c.doRequest(ctx, http.MethodGet, usersURL.String(), nil, &usersResponse)
	if err != nil {
		return nil, "", nil, err
	}

	nextToken := getNextToken(usersResponse.CurrentPage, usersResponse.TotalPages)

	var users []*ZuperUser
	for _, user := range usersResponse.Data {
		users = append(users, &user)
	}

	return users, nextToken, annos, nil
}

// GetUserByID fetches the details of a user by their user_uid from the Zuper API.
func (c *Client) GetUserByID(ctx context.Context, userUID string) (*ZuperUser, annotations.Annotations, error) {
	userURL, err := buildResourceURL(c.apiUrl, userEndpoint, userUID)
	if err != nil {
		return nil, nil, err
	}
	var userResponse UserDetailsResponse
	_, annos, err := c.doRequest(ctx, http.MethodGet, userURL, nil, &userResponse)
	if err != nil {
		return nil, nil, err
	}
	return &userResponse.Data, annos, nil
}

// CreateUser sends a request to create a new user with the provided user payload and default work hours.
func (c *Client) CreateUser(ctx context.Context, user UserPayload) (*CreateUserResponse, annotations.Annotations, error) {
	workHours := []WorkHour{
		{Day: "Sunday", StartTime: "06:00 AM", EndTime: "06:00 PM", WorkMins: 0, TrackLocation: true, IsEnabled: "false"},
		{Day: "Monday", StartTime: "06:00 AM", EndTime: "06:00 PM", WorkMins: 0, TrackLocation: true, IsEnabled: "false"},
		{Day: "Tuesday", StartTime: "06:00 AM", EndTime: "06:00 PM", WorkMins: 0, TrackLocation: true, IsEnabled: "false"},
		{Day: "Wednesday", StartTime: "06:00 AM", EndTime: "06:00 PM", WorkMins: 0, TrackLocation: true, IsEnabled: "false"},
		{Day: "Thursday", StartTime: "06:00 AM", EndTime: "06:00 PM", WorkMins: 0, TrackLocation: true, IsEnabled: "false"},
		{Day: "Friday", StartTime: "06:00 AM", EndTime: "06:00 PM", WorkMins: 0, TrackLocation: true, IsEnabled: "false"},
		{Day: "Saturday", StartTime: "06:00 AM", EndTime: "06:00 PM", WorkMins: 0, TrackLocation: true, IsEnabled: "false"},
	}

	payload := CreateUserRequest{
		WorkHours: workHours,
		User:      user,
	}

	userCreateURL, err := buildResourceURL(c.apiUrl, userEndpoint)
	if err != nil {
		return nil, nil, err
	}
	var result CreateUserResponse
	_, annos, err := c.doRequest(ctx, http.MethodPost, userCreateURL, payload, &result)
	if err != nil {
		return nil, nil, err
	}

	return &result, annos, nil
}

// doRequest executes an HTTP request and decodes the response into the provided result. If body is not nil, it serializes it and sends it as the body.
func (c *Client) doRequest(
	ctx context.Context,
	method string,
	requestURL string,
	body interface{},
	res interface{},
) (http.Header, annotations.Annotations, error) {
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return nil, nil, err
	}

	var zuperErr ZuperError
	requestOptions := []uhttp.RequestOption{
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithHeader("x-api-key", c.apiKey),
	}
	if body != nil {
		requestOptions = append(requestOptions, uhttp.WithJSONBody(body))
	}

	req, err := c.wrapper.NewRequest(
		ctx,
		method,
		parsedURL,
		requestOptions...,
	)
	if err != nil {
		return nil, nil, err
	}

	var doOptions []uhttp.DoOption
	if res != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(res))
	}
	doOptions = append(doOptions, uhttp.WithErrorResponse(&zuperErr))

	resp, err := c.wrapper.Do(req, doOptions...)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	annos := annotations.Annotations{}
	if desc, err := ratelimit.ExtractRateLimitData(resp.StatusCode, &resp.Header); err == nil {
		annos.WithRateLimiting(desc)
	}

	return resp.Header, annos, nil
}
