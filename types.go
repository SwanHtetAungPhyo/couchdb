package couchdb

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"time"
)

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
