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
	getUsers = "/api/user/all"
)

type Client struct {
	apiUrl  string
	Token   string
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
		Token:   client.Token,
	}, nil
}

func NewClient(ctx context.Context, apiUrl string, token string, httpClient *uhttp.BaseHttpClient) *Client {
	if httpClient == nil {
		httpClient = &uhttp.BaseHttpClient{}
	}
	return &Client{
		wrapper: httpClient,
		apiUrl:  apiUrl,
		Token:   token,
	}
}

func (c *Client) GetUsers(ctx context.Context, pToken string) ([]*ZuperUser, string, annotations.Annotations, error) {
	opts := pageOptions{
		PageToken: pToken,
		PageSize:  defaultPageSize,
	}

	usersURL, _, err := preparePagedRequest(c.apiUrl, getUsers, opts)
	if err != nil {
		return nil, "", nil, err
	}

	var usersResponse UsersResponse
	_, annos, err := c.doRequest(ctx, http.MethodGet, usersURL.String(), &usersResponse)
	if err != nil {
		return nil, "", nil, err
	}

	nextToken := getNextToken(usersResponse.CurrentPage, usersResponse.TotalPages)

	// Convert to pointers
	var users []*ZuperUser
	for i := range usersResponse.Data {
		users = append(users, &usersResponse.Data[i])
	}

	return users, nextToken, annos, nil
}

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
	var zuperErr ZuperError
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
		doOptions = append(doOptions, uhttp.WithErrorResponse(&zuperErr))
		resp, err = c.wrapper.Do(req, doOptions...)
	case http.MethodDelete:
		resp, err = c.wrapper.Do(req)
	}

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
