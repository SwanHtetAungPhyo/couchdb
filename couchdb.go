// Package couchdb provides a Go client library for Apache CouchDB using Resty
package couchdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// NewClient creates a new CouchDB client
func NewClient(baseURL string, opts *ClientOptions) *Client {
	if opts == nil {
		opts = &ClientOptions{}
	}

	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	client := resty.New()
	client.SetBaseURL(strings.TrimSuffix(baseURL, "/"))
	client.SetTimeout(opts.Timeout)
	client.SetHeader("Content-Type", "application/json")
	client.SetDebug(opts.Debug)

	if opts.Username != "" && opts.Password != "" {
		client.SetBasicAuth(opts.Username, opts.Password)
	}

	return &Client{
		resty:   client,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

type ServerInfo struct {
	CouchDB string `json:"couchdb"`
	Version string `json:"version"`
	UUID    string `json:"uuid"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("CouchDB error %d: %s - %s", e.StatusCode, e.Type, e.Reason)
}

// CompactDesignDoc compacts a specific design document's view indexes
func (db *Database) CompactDesignDoc(ctx context.Context, designDoc string) error {
	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		Post("/" + db.name + "/_compact/" + designDoc)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return db.client.parseError(resp)
	}

	return nil
}

// Helper methods for common view patterns

// ViewByKey is a convenience method to query a view by a single key
func (db *Database) ViewByKey(ctx context.Context, designDoc, viewName string, key interface{}) (*ViewResult, error) {
	return db.View(ctx, designDoc, viewName, &ViewOptions{
		Key: key,
	})
}

// ViewByKeyRange is a convenience method to query a view by key range
func (db *Database) ViewByKeyRange(ctx context.Context, designDoc, viewName string, startKey, endKey interface{}) (*ViewResult, error) {
	return db.View(ctx, designDoc, viewName, &ViewOptions{
		StartKey: startKey,
		EndKey:   endKey,
	})
}

// ViewAll is a convenience method to get all results from a view
func (db *Database) ViewAll(ctx context.Context, designDoc, viewName string, includeDocs bool) (*ViewResult, error) {
	return db.View(ctx, designDoc, viewName, &ViewOptions{
		IncludeDocs: includeDocs,
	})
}

// Changes returns database changes
func (db *Database) Changes(ctx context.Context, opts map[string]interface{}) (map[string]interface{}, error) {
	req := db.client.resty.R().SetContext(ctx)

	for k, v := range opts {
		req.SetQueryParam(k, fmt.Sprintf("%v", v))
	}

	var result map[string]interface{}
	resp, err := req.
		SetResult(&result).
		Get("/" + db.name + "/_changes")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return result, nil
}

// Compact triggers database compaction
func (db *Database) Compact(ctx context.Context) error {
	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		Post("/" + db.name + "/_compact")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return db.client.parseError(resp)
	}

	return nil
}
