package couchdb

import (
	"context"
	"fmt"
)

// ViewReduce is a convenience method to get reduced results from a view
func (db *Database) ViewReduce(ctx context.Context, designDoc, viewName string, groupLevel int) (*ViewResult, error) {
	reduce := true
	opts := &ViewOptions{
		Reduce: &reduce,
	}

	if groupLevel > 0 {
		opts.GroupLevel = groupLevel
	} else {
		opts.Group = true
	}

	return db.View(ctx, designDoc, viewName, opts)
}

// Get retrieves a document by ID
func (db *Database) Get(ctx context.Context, id string, rev ...string) (*Document, error) {
	req := db.client.resty.R().SetContext(ctx)

	if len(rev) > 0 && rev[0] != "" {
		req.SetQueryParam("rev", rev[0])
	}

	var doc Document
	resp, err := req.
		SetResult(&doc).
		Get("/" + db.name + "/" + id)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return &doc, nil
}

// Put creates or updates a document
func (db *Database) Put(ctx context.Context, doc interface{}) (*Document, error) {
	var result struct {
		ID  string `json:"id"`
		Rev string `json:"rev"`
		OK  bool   `json:"ok"`
	}

	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetBody(doc).
		SetResult(&result).
		Post("/" + db.name)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return &Document{ID: result.ID, Rev: result.Rev}, nil
}

// Update updates a document with a specific ID
func (db *Database) Update(ctx context.Context, id string, doc interface{}) (*Document, error) {
	var result struct {
		ID  string `json:"id"`
		Rev string `json:"rev"`
		OK  bool   `json:"ok"`
	}

	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetBody(doc).
		SetResult(&result).
		Put("/" + db.name + "/" + id)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return &Document{ID: result.ID, Rev: result.Rev}, nil
}

// Delete deletes a document
func (db *Database) Delete(ctx context.Context, id, rev string) error {
	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetQueryParam("rev", rev).
		Delete("/" + db.name + "/" + id)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return db.client.parseError(resp)
	}

	return nil
}

// AllDocs retrieves all documents
func (db *Database) AllDocs(ctx context.Context, opts *ViewOptions) (*ViewResult, error) {
	req := db.client.resty.R().SetContext(ctx)

	if opts != nil {
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
		if opts.StartKey != nil {
			req.SetQueryParam("startkey", fmt.Sprintf("%v", opts.StartKey))
		}
		if opts.EndKey != nil {
			req.SetQueryParam("endkey", fmt.Sprintf("%v", opts.EndKey))
		}
	}

	var result ViewResult
	resp, err := req.
		SetResult(&result).
		Get("/" + db.name + "/_all_docs")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return &result, nil
}

// Bulk performs bulk operations
func (db *Database) Bulk(ctx context.Context, docs []interface{}) ([]BulkResult, error) {
	bulkDocs := BulkDocs{
		Docs: docs,
	}

	var results []BulkResult
	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetBody(bulkDocs).
		SetResult(&results).
		Post("/" + db.name + "/_bulk_docs")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return results, nil
}
