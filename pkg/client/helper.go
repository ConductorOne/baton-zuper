package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

const DefaultPageSize = 50

// EncodePageToken serializes the pageToken to base64.
func EncodePageToken(pt *pageToken) string {
	b, _ := json.Marshal(pt)
	return base64.StdEncoding.EncodeToString(b)
}

// DecodePageToken deserializes a base64 token to pageToken.
func DecodePageToken(token string) (*pageToken, error) {
	if token == "" {
		return &pageToken{Page: 1}, nil
	}
	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var pt pageToken
	if err := json.Unmarshal(data, &pt); err != nil {
		return nil, err
	}
	return &pt, nil
}

// PreparePagedRequest prepares the URL with pagination parameters.
func PreparePagedRequest(baseURL, endpoint string, opts pageOptions) (*url.URL, int, error) {
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
		pt, err := DecodePageToken(opts.PageToken)
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

// GetNextToken generates the token for the next page.
func GetNextToken(current, total int) string {
	if current < total {
		return EncodePageToken(&pageToken{Page: current + 1})
	}
	return ""
}
