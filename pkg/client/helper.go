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

// buildResourceURL builds a request URL for the Zuper API for any endpoint and optional path elements.
func buildResourceURL(baseURL string, endpoint string, elems ...string) (string, error) {
	joined, err := url.JoinPath(baseURL, append([]string{endpoint}, elems...)...)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	return joined, nil
}

// preparePagedRequest builds a paginated request URL for the Zuper API.
func preparePagedRequest(baseURL string, endpoint string, opts PageOptions, elems ...string) (*url.URL, int, error) {
	urlStr, err := buildResourceURL(baseURL, endpoint, elems...)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid base or endpoint: %w", err)
	}
	fullURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid URL: %w", err)
	}

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
func getNextToken(current int, total int) string {
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
