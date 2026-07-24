package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type HostIMGClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewHostIMGClient(baseURL, token string) *HostIMGClient {
	return &HostIMGClient{baseURL: baseURL, token: token, http: &http.Client{}}
}

func (c *HostIMGClient) Do(ctx context.Context, method, path string, userID int64, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hostimg request failed: %w", err)
	}
	return resp, nil
}
