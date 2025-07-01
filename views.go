package couchdb

import (
	"context"
	"encoding/json"
	"fmt"
)

// Enhanced View Methods

// View executes a view query with comprehensive options
func (db *Database) View(ctx context.Context, designDoc, viewName string, opts *ViewOptions) (*ViewResult, error) {
	req := db.client.resty.R().SetContext(ctx)

	if opts != nil {
		// Handle key-based queries
		if opts.Key != nil {
			keyBytes, _ := json.Marshal(opts.Key)
			req.SetQueryParam("key", string(keyBytes))
		}

		if len(opts.Keys) > 0 {
			keysBytes, _ := json.Marshal(opts.Keys)
			req.SetQueryParam("keys", string(keysBytes))
		}

		if opts.StartKey != nil {
			keyBytes, _ := json.Marshal(opts.StartKey)
			req.SetQueryParam("startkey", string(keyBytes))
		}

		if opts.EndKey != nil {
			keyBytes, _ := json.Marshal(opts.EndKey)
			req.SetQueryParam("endkey", string(keyBytes))
		}

		if opts.StartKeyDocID != "" {
			req.SetQueryParam("startkey_docid", opts.StartKeyDocID)
		}

		if opts.EndKeyDocID != "" {
			req.SetQueryParam("endkey_docid", opts.EndKeyDocID)
		}

		// Result control
		if opts.Limit > 0 {
			req.SetQueryParam("limit", fmt.Sprintf("%d", opts.Limit))
		}

		if opts.Skip > 0 {
			req.SetQueryParam("skip", fmt.Sprintf("%d", opts.Skip))
		}

		if opts.Descending {
			req.SetQueryParam("descending", "true")
		}

		if opts.InclusiveEnd {
			req.SetQueryParam("inclusive_end", "true")
		}

		// Group/Reduce options
		if opts.Group {
			req.SetQueryParam("group", "true")
		}

		if opts.GroupLevel > 0 {
			req.SetQueryParam("group_level", fmt.Sprintf("%d", opts.GroupLevel))
		}

		if opts.Reduce != nil {
			req.SetQueryParam("reduce", fmt.Sprintf("%t", *opts.Reduce))
		}

		// Additional options
		if opts.IncludeDocs {
			req.SetQueryParam("include_docs", "true")
		}

		if opts.UpdateSeq {
			req.SetQueryParam("update_seq", "true")
		}

		if opts.Conflicts {
			req.SetQueryParam("conflicts", "true")
		}

		if opts.Attachments {
			req.SetQueryParam("attachments", "true")
		}

		if opts.AttEncodingInfo {
			req.SetQueryParam("att_encoding_info", "true")
		}

		// Staleness control
		if opts.Stale != "" {
			req.SetQueryParam("stale", opts.Stale)
		}

		if opts.Update != "" {
			req.SetQueryParam("update", opts.Update)
		}
	}

	var result ViewResult
	resp, err := req.
		SetResult(&result).
		Get("/" + db.name + "/_design/" + designDoc + "/_view/" + viewName)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return &result, nil
}

// ViewWithKeys executes a view query with multiple keys (POST request)
func (db *Database) ViewWithKeys(ctx context.Context, designDoc, viewName string, keys []interface{}, opts *ViewOptions) (*ViewResult, error) {
	body := map[string]interface{}{
		"keys": keys,
	}

	req := db.client.resty.R().
		SetContext(ctx).
		SetBody(body)

	if opts != nil {
		// Add query parameters for other options
		if opts.IncludeDocs {
			req.SetQueryParam("include_docs", "true")
		}
		if opts.Limit > 0 {
			req.SetQueryParam("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Skip > 0 {
			req.SetQueryParam("skip", fmt.Sprintf("%d", opts.Skip))
		}
		if opts.Descending {
			req.SetQueryParam("descending", "true")
		}
		if opts.Group {
			req.SetQueryParam("group", "true")
		}
		if opts.GroupLevel > 0 {
			req.SetQueryParam("group_level", fmt.Sprintf("%d", opts.GroupLevel))
		}
		if opts.Reduce != nil {
			req.SetQueryParam("reduce", fmt.Sprintf("%t", *opts.Reduce))
		}
	}

	var result ViewResult
	resp, err := req.
		SetResult(&result).
		Post("/" + db.name + "/_design/" + designDoc + "/_view/" + viewName)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return &result, nil
}

// ViewInfo gets information about a view
func (db *Database) ViewInfo(ctx context.Context, designDoc, viewName string) (map[string]interface{}, error) {
	var result map[string]interface{}
	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/" + db.name + "/_design/" + designDoc + "/_info")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return result, nil
}

// ViewCleanup removes old view index files
func (db *Database) ViewCleanup(ctx context.Context) error {
	resp, err := db.client.resty.R().
		SetContext(ctx).
		Post("/" + db.name + "/_view_cleanup")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return db.client.parseError(resp)
	}

	return nil
}
