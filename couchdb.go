// Package couchdb provides a Go client library for Apache CouchDB using Resty
package couchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client represents a CouchDB client
type Client struct {
	resty   *resty.Client
	baseURL string
}

// ClientOptions holds configuration options for the CouchDB client
type ClientOptions struct {
	Username string
	Password string
	Timeout  time.Duration
	Debug    bool
}

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

// Database represents a CouchDB database
type Database struct {
	client *Client
	name   string
}

// DB returns a Database instance for the specified database name
func (c *Client) DB(name string) *Database {
	return &Database{
		client: c,
		name:   name,
	}
}

// Document represents a CouchDB document
type Document struct {
	ID      string                 `json:"_id,omitempty"`
	Rev     string                 `json:"_rev,omitempty"`
	Deleted bool                   `json:"_deleted,omitempty"`
	Data    map[string]interface{} `json:"-"`
}

// MarshalJSON implements json.Marshaler
func (d *Document) MarshalJSON() ([]byte, error) {
	doc := make(map[string]interface{})

	// Copy data fields
	for k, v := range d.Data {
		doc[k] = v
	}

	// Add system fields
	if d.ID != "" {
		doc["_id"] = d.ID
	}
	if d.Rev != "" {
		doc["_rev"] = d.Rev
	}
	if d.Deleted {
		doc["_deleted"] = d.Deleted
	}

	return json.Marshal(doc)
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Document) UnmarshalJSON(data []byte) error {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}

	d.Data = make(map[string]interface{})

	for k, v := range doc {
		switch k {
		case "_id":
			if id, ok := v.(string); ok {
				d.ID = id
			}
		case "_rev":
			if rev, ok := v.(string); ok {
				d.Rev = rev
			}
		case "_deleted":
			if deleted, ok := v.(bool); ok {
				d.Deleted = deleted
			}
		default:
			d.Data[k] = v
		}
	}

	return nil
}

// DesignDocument represents a CouchDB design document
type DesignDocument struct {
	ID       string            `json:"_id,omitempty"`
	Rev      string            `json:"_rev,omitempty"`
	Language string            `json:"language,omitempty"`
	Views    map[string]*View  `json:"views,omitempty"`
	Shows    map[string]string `json:"shows,omitempty"`
	Lists    map[string]string `json:"lists,omitempty"`
	Updates  map[string]string `json:"updates,omitempty"`
	Filters  map[string]string `json:"filters,omitempty"`
	Validate string            `json:"validate_doc_update,omitempty"`
}

// View represents a CouchDB view with map and reduce functions
type View struct {
	Map    string `json:"map"`
	Reduce string `json:"reduce,omitempty"`
}

// ViewResult represents the result of a view query
type ViewResult struct {
	TotalRows int64     `json:"total_rows"`
	Offset    int64     `json:"offset"`
	Rows      []ViewRow `json:"rows"`
	UpdateSeq string    `json:"update_seq,omitempty"`
}

// ViewRow represents a single row in a view result
type ViewRow struct {
	ID    string      `json:"id"`
	Key   interface{} `json:"key"`
	Value interface{} `json:"value"`
	Doc   *Document   `json:"doc,omitempty"`
}

// ViewOptions holds options for view queries
type ViewOptions struct {
	// Key selection
	Key           interface{}   `json:"key,omitempty"`
	Keys          []interface{} `json:"keys,omitempty"`
	StartKey      interface{}   `json:"startkey,omitempty"`
	EndKey        interface{}   `json:"endkey,omitempty"`
	StartKeyDocID string        `json:"startkey_docid,omitempty"`
	EndKeyDocID   string        `json:"endkey_docid,omitempty"`

	// Result control
	Limit        int  `json:"limit,omitempty"`
	Skip         int  `json:"skip,omitempty"`
	Descending   bool `json:"descending,omitempty"`
	InclusiveEnd bool `json:"inclusive_end,omitempty"`

	// Group/Reduce
	Group      bool  `json:"group,omitempty"`
	GroupLevel int   `json:"group_level,omitempty"`
	Reduce     *bool `json:"reduce,omitempty"`

	// Additional options
	IncludeDocs     bool `json:"include_docs,omitempty"`
	UpdateSeq       bool `json:"update_seq,omitempty"`
	Conflicts       bool `json:"conflicts,omitempty"`
	Attachments     bool `json:"attachments,omitempty"`
	AttEncodingInfo bool `json:"att_encoding_info,omitempty"`

	// Staleness
	Stale  string `json:"stale,omitempty"`  // "ok" or "update_after"
	Update string `json:"update,omitempty"` // "true", "false", or "lazy"
}

// ViewQuery represents a structured view query
type ViewQuery struct {
	DesignDoc string
	ViewName  string
	Options   *ViewOptions
}

// ViewBuilder helps build complex view queries
type ViewBuilder struct {
	designDoc string
	viewName  string
	options   *ViewOptions
}

// NewViewQuery creates a new view query builder
func (db *Database) NewViewQuery(designDoc, viewName string) *ViewBuilder {
	return &ViewBuilder{
		designDoc: designDoc,
		viewName:  viewName,
		options:   &ViewOptions{},
	}
}

// Key sets a specific key to query
func (vb *ViewBuilder) Key(key interface{}) *ViewBuilder {
	vb.options.Key = key
	return vb
}

// Keys sets multiple keys to query
func (vb *ViewBuilder) Keys(keys ...interface{}) *ViewBuilder {
	vb.options.Keys = keys
	return vb
}

// StartKey sets the start key for range queries
func (vb *ViewBuilder) StartKey(key interface{}) *ViewBuilder {
	vb.options.StartKey = key
	return vb
}

// EndKey sets the end key for range queries
func (vb *ViewBuilder) EndKey(key interface{}) *ViewBuilder {
	vb.options.EndKey = key
	return vb
}

// Limit sets the maximum number of results
func (vb *ViewBuilder) Limit(limit int) *ViewBuilder {
	vb.options.Limit = limit
	return vb
}

// Skip sets the number of results to skip
func (vb *ViewBuilder) Skip(skip int) *ViewBuilder {
	vb.options.Skip = skip
	return vb
}

// Descending sets the sort order
func (vb *ViewBuilder) Descending(desc bool) *ViewBuilder {
	vb.options.Descending = desc
	return vb
}

// Group enables grouping for reduce views
func (vb *ViewBuilder) Group(group bool) *ViewBuilder {
	vb.options.Group = group
	return vb
}

// GroupLevel sets the group level for reduce views
func (vb *ViewBuilder) GroupLevel(level int) *ViewBuilder {
	vb.options.GroupLevel = level
	return vb
}

// Reduce controls whether to use reduce function
func (vb *ViewBuilder) Reduce(reduce bool) *ViewBuilder {
	vb.options.Reduce = &reduce
	return vb
}

// IncludeDocs includes the full document in results
func (vb *ViewBuilder) IncludeDocs(include bool) *ViewBuilder {
	vb.options.IncludeDocs = include
	return vb
}

// Stale controls staleness tolerance
func (vb *ViewBuilder) Stale(stale string) *ViewBuilder {
	vb.options.Stale = stale
	return vb
}

// Execute runs the view query
func (vb *ViewBuilder) Execute(ctx context.Context, db *Database) (*ViewResult, error) {
	return db.View(ctx, vb.designDoc, vb.viewName, vb.options)
}

type ServerInfo struct {
	CouchDB string `json:"couchdb"`
	Version string `json:"version"`
	UUID    string `json:"uuid"`
}

type DatabaseInfo struct {
	DBName            string `json:"db_name"`
	DocCount          int64  `json:"doc_count"`
	DocDelCount       int64  `json:"doc_del_count"`
	UpdateSeq         string `json:"update_seq"`
	PurgeSeq          string `json:"purge_seq"`
	CompactRunning    bool   `json:"compact_running"`
	DiskSize          int64  `json:"disk_size"`
	DataSize          int64  `json:"data_size"`
	InstanceStartTime string `json:"instance_start_time"`
}

type BulkResult struct {
	ID     string `json:"id"`
	Rev    string `json:"rev,omitempty"`
	Error  string `json:"error,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type BulkDocs struct {
	Docs     []interface{} `json:"docs"`
	NewEdits bool          `json:"new_edits,omitempty"`
}

type Error struct {
	StatusCode int    `json:"-"`
	Type       string `json:"error"`
	Reason     string `json:"reason"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("CouchDB error %d: %s - %s", e.StatusCode, e.Type, e.Reason)
}

// Design Document Methods

// GetDesignDoc retrieves a design document
func (db *Database) GetDesignDoc(ctx context.Context, name string) (*DesignDocument, error) {
	var designDoc DesignDocument
	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetResult(&designDoc).
		Get("/" + db.name + "/_design/" + name)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return &designDoc, nil
}

// PutDesignDoc creates or updates a design document
func (db *Database) PutDesignDoc(ctx context.Context, name string, designDoc *DesignDocument) (*Document, error) {
	if designDoc.ID == "" {
		designDoc.ID = "_design/" + name
	}
	if designDoc.Language == "" {
		designDoc.Language = "javascript"
	}

	var result struct {
		ID  string `json:"id"`
		Rev string `json:"rev"`
		OK  bool   `json:"ok"`
	}

	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetBody(designDoc).
		SetResult(&result).
		Put("/" + db.name + "/_design/" + name)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return &Document{ID: result.ID, Rev: result.Rev}, nil
}

// DeleteDesignDoc deletes a design document
func (db *Database) DeleteDesignDoc(ctx context.Context, name, rev string) error {
	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetQueryParam("rev", rev).
		Delete("/" + db.name + "/_design/" + name)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return db.client.parseError(resp)
	}

	return nil
}

// ListDesignDocs lists all design documents
func (db *Database) ListDesignDocs(ctx context.Context) (*ViewResult, error) {
	var result ViewResult
	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetQueryParam("startkey", `"_design/"`).
		SetQueryParam("endkey", `"_design0"`).
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

// Client methods

// Info returns server information
func (c *Client) Info(ctx context.Context) (*ServerInfo, error) {
	var info ServerInfo
	resp, err := c.resty.R().
		SetContext(ctx).
		SetResult(&info).
		Get("/")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, c.parseError(resp)
	}

	return &info, nil
}

// AllDbs returns a list of all databases
func (c *Client) AllDbs(ctx context.Context) ([]string, error) {
	var dbs []string
	resp, err := c.resty.R().
		SetContext(ctx).
		SetResult(&dbs).
		Get("/_all_dbs")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, c.parseError(resp)
	}

	return dbs, nil
}

// CreateDB creates a new database
func (c *Client) CreateDB(ctx context.Context, name string) error {
	resp, err := c.resty.R().
		SetContext(ctx).
		Put("/" + name)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return c.parseError(resp)
	}

	return nil
}

// DeleteDB deletes a database
func (c *Client) DeleteDB(ctx context.Context, name string) error {
	resp, err := c.resty.R().
		SetContext(ctx).
		Delete("/" + name)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return c.parseError(resp)
	}

	return nil
}

// Database methods

// Info returns database information
func (db *Database) Info(ctx context.Context) (*DatabaseInfo, error) {
	var info DatabaseInfo
	resp, err := db.client.resty.R().
		SetContext(ctx).
		SetResult(&info).
		Get("/" + db.name)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, db.client.parseError(resp)
	}

	return &info, nil
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
