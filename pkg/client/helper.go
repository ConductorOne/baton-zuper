package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

const DefaultPageSize = 50

type ErrorResponse interface {
	Message() string
}

// encodePageToken encodes a pageToken struct into a base64 string.
func encodePageToken(pToken *pageToken) (string, error) {
	b, err := json.Marshal(pToken)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// decodePageToken decodes a base64 string into a pageToken struct.
func decodePageToken(pToken string) (*pageToken, error) {
	if pToken == "" {
		return &pageToken{Page: 1}, nil
	}
	data, err := base64.StdEncoding.DecodeString(pToken)
	if err != nil {
		return nil, err
	}
	var pt pageToken
	if err := json.Unmarshal(data, &pt); err != nil {
		return nil, err
	}
	return &pt, nil
}

// preparePagedRequest builds a paginated request URL for the Zuper API.
func preparePagedRequest(baseURL, endpoint string, opts PageOptions) (*url.URL, int, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid base URL: %w", err)
	}

	rel, err := url.Parse(endpoint)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid endpoint: %w", err)
	}

	fullURL := base.ResolveReference(rel)

	page := 1
	if opts.PageToken != "" {
		pt, err := decodePageToken(opts.PageToken)
		if err != nil {
			return nil, 0, err
		}
		page = pt.Page
	}

	q := fullURL.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", fmt.Sprintf("%d", opts.PageSize))

	fullURL.RawQuery = q.Encode()

	return fullURL, page, nil
}

// getNextToken returns the next page token if more pages are available.
func getNextToken(current, total int) string {
	if current < total {
		token, err := encodePageToken(&pageToken{Page: current + 1})
		if err != nil {
			return ""
		}
		return token
	}
	return ""
}

// Error implements the uhttp.ErrorResponse interface.
func (e *ZuperError) Message() string {
	return e.MessageError
}

// prepareUserDetailsRequest builds a request URL for the Zuper API user details endpoint.
func prepareUserDetailsRequest(baseURL string, endpoint string, userUID string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	rel, err := url.Parse(endpoint + userUID)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint: %w", err)
	}

	fullURL := base.ResolveReference(rel)
	return fullURL.String(), nil
}

// prepareUserCreateRequest builds a request URL for the Zuper API user create endpoint.
func prepareUserCreateRequest(baseURL string, endpoint string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	rel, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint: %w", err)
	}

	fullURL := base.ResolveReference(rel)
	return fullURL.String(), nil
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(fmt.Sprintf("invalid URL: %s", raw))
	}
	return u
}
