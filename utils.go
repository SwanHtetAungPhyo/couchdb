package couchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
)

// Helper methods

func (c *Client) parseError(resp *resty.Response) error {
	var couchError Error
	couchError.StatusCode = resp.StatusCode()

	if err := json.Unmarshal(resp.Body(), &couchError); err != nil {
		couchError.Type = "unknown"
		couchError.Reason = string(resp.Body())
	}

	return &couchError
}

// Utility functions

// UUID generates a UUID from CouchDB
func (c *Client) UUID(ctx context.Context) (string, error) {
	var result struct {
		UUIDs []string `json:"uuids"`
	}

	resp, err := c.resty.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/_uuids")

	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", c.parseError(resp)
	}

	if len(result.UUIDs) > 0 {
		return result.UUIDs[0], nil
	}

	return "", fmt.Errorf("no UUID returned")
}

// UUIDs generates multiple UUIDs from CouchDB
func (c *Client) UUIDs(ctx context.Context, count int) ([]string, error) {
	var result struct {
		UUIDs []string `json:"uuids"`
	}

	resp, err := c.resty.R().
		SetContext(ctx).
		SetQueryParam("count", fmt.Sprintf("%d", count)).
		SetResult(&result).
		Get("/_uuids")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, c.parseError(resp)
	}

	return result.UUIDs, nil
}
