package couchdb

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestSuite handles both mock and real CouchDB testing
type CouchDBTestSuite struct {
	suite.Suite
	client     *Client
	testDB     *Database
	useMock    bool
	mockServer *httptest.Server
}

// SetupSuite runs before all tests
func (suite *CouchDBTestSuite) SetupSuite() {
	couchURL := os.Getenv("COUCHDB_URL")
	couchUser := os.Getenv("COUCHDB_USER")
	couchPass := os.Getenv("COUCHDB_PASS")

	if couchURL != "" {
		// Use real CouchDB instance
		suite.useMock = false
		opts := &ClientOptions{}
		if couchUser != "" && couchPass != "" {
			opts.Username = couchUser
			opts.Password = couchPass
		}
		suite.client = NewClient(couchURL, opts)
	} else {
		suite.useMock = true
		suite.setupMockServer()
	}

	suite.testDB = suite.client.DB("test-db-" + time.Now().Format("20060102150405"))
}

// TearDownSuite runs after all tests
func (suite *CouchDBTestSuite) TearDownSuite() {
	if !suite.useMock && suite.testDB != nil {
		err := suite.client.DeleteDB(context.Background(), suite.testDB.name)
		if err != nil {
			return
		}
	}
	if suite.mockServer != nil {
		suite.mockServer.Close()
	}
}

func (suite *CouchDBTestSuite) setupMockServer() {
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.handleMockRequest(w, r)
	}))
	suite.client = NewClient(suite.mockServer.URL, nil)
}

func (suite *CouchDBTestSuite) handleMockRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == "GET" && r.URL.Path == "/":
		err := json.NewEncoder(w).Encode(ServerInfo{
			CouchDB: "Welcome",
			Version: "3.2.0",
			UUID:    "mock-uuid-123",
		})
		if err != nil {
			return

		}
	case r.Method == "GET" && r.URL.Path == "/_all_dbs":
		err := json.NewEncoder(w).Encode([]string{"_users", "_replicator", "test-db"})
		if err != nil {
			return
		}
	case r.Method == "PUT" && r.URL.Path == "/"+suite.testDB.name:
		w.WriteHeader(http.StatusCreated)
		err := json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
		if err != nil {
			return
		}
	case r.Method == "GET" && r.URL.Path == "/"+suite.testDB.name:
		err := json.NewEncoder(w).Encode(DatabaseInfo{
			DBName:    suite.testDB.name,
			DocCount:  0,
			UpdateSeq: "0",
		})
		if err != nil {
			return
		}
	case r.Method == "POST" && r.URL.Path == "/"+suite.testDB.name:
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":  true,
			"id":  "mock-doc-id",
			"rev": "1-mock-rev",
		})
		if err != nil {
			return
		}
	case r.Method == "GET" && r.URL.Path == "/"+suite.testDB.name+"/test-doc":
		err := json.NewEncoder(w).Encode(Document{
			ID:  "test-doc",
			Rev: "1-mock-rev",
			Data: map[string]interface{}{
				"name": "Test Document",
				"type": "test",
			},
		})
		if err != nil {
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  "not_found",
			"reason": "missing",
		})
		if err != nil {
			return
		}
	}
}

// Test NewClient
func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		opts     *ClientOptions
		expected string
	}{
		{
			name:     "basic client creation",
			baseURL:  "http://localhost:5984",
			opts:     nil,
			expected: "http://localhost:5984",
		},
		{
			name:     "client with trailing slash",
			baseURL:  "http://localhost:5984/",
			opts:     nil,
			expected: "http://localhost:5984",
		},
		{
			name:    "client with options",
			baseURL: "http://localhost:5984",
			opts: &ClientOptions{
				Username: "admin",
				Password: "admin_password",
				Timeout:  10 * time.Second,
				Debug:    true,
			},
			expected: "http://localhost:5984",
		},
		{
			name:     "client with default timeout",
			baseURL:  "http://localhost:5984",
			opts:     &ClientOptions{},
			expected: "http://localhost:5984",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.opts)
			assert.NotNil(t, client)
			assert.Equal(t, tt.expected, client.baseURL)
			assert.NotNil(t, client.resty)
		})
	}
}

// Test Document JSON marshaling/unmarshalling
func TestDocument_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		doc      *Document
		expected map[string]interface{}
	}{
		{
			name: "document with all fields",
			doc: &Document{
				ID:      "test-id",
				Rev:     "1-abc123",
				Deleted: false,
				Data: map[string]interface{}{
					"name": "test",
					"age":  30,
				},
			},
			expected: map[string]interface{}{
				"_id":  "test-id",
				"_rev": "1-abc123",
				"name": "test",
				"age":  float64(30),
			},
		},
		{
			name: "document with deleted flag",
			doc: &Document{
				ID:      "test-id",
				Rev:     "2-def456",
				Deleted: true,
				Data:    map[string]interface{}{},
			},
			expected: map[string]interface{}{
				"_id":      "test-id",
				"_rev":     "2-def456",
				"_deleted": true,
			},
		},
		{
			name: "document with only data",
			doc: &Document{
				Data: map[string]interface{}{
					"title": "Test Document",
					"value": 42,
				},
			},
			expected: map[string]interface{}{
				"title": "Test Document",
				"value": float64(42),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.doc.MarshalJSON()
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDocument_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Document
	}{
		{
			name:  "full document",
			input: `{"_id":"test-id","_rev":"1-abc123","name":"test","age":30}`,
			expected: &Document{
				ID:  "test-id",
				Rev: "1-abc123",
				Data: map[string]interface{}{
					"name": "test",
					"age":  float64(30),
				},
			},
		},
		{
			name:  "deleted document",
			input: `{"_id":"test-id","_rev":"2-def456","_deleted":true}`,
			expected: &Document{
				ID:      "test-id",
				Rev:     "2-def456",
				Deleted: true,
				Data:    map[string]interface{}{},
			},
		},
		{
			name:  "document without system fields",
			input: `{"title":"Test","value":42}`,
			expected: &Document{
				Data: map[string]interface{}{
					"title": "Test",
					"value": float64(42),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc Document
			err := doc.UnmarshalJSON([]byte(tt.input))
			require.NoError(t, err)

			assert.Equal(t, tt.expected.ID, doc.ID)
			assert.Equal(t, tt.expected.Rev, doc.Rev)
			assert.Equal(t, tt.expected.Deleted, doc.Deleted)
			assert.Equal(t, tt.expected.Data, doc.Data)
		})
	}
}

// Test ViewBuilder
func TestViewBuilder(t *testing.T) {
	client := NewClient("http://localhost:5984", nil)
	db := client.DB("test-db")

	vb := db.NewViewQuery("test-design", "test-view")
	assert.NotNil(t, vb)
	assert.Equal(t, "test-design", vb.designDoc)
	assert.Equal(t, "test-view", vb.viewName)
	assert.NotNil(t, vb.options)

	// Test chaining
	vb = vb.Key("test-key").
		StartKey("start").
		EndKey("end").
		Limit(10).
		Skip(5).
		Descending(true).
		Group(true).
		GroupLevel(2).
		Reduce(false).
		IncludeDocs(true).
		Stale("ok")

	assert.Equal(t, "test-key", vb.options.Key)
	assert.Equal(t, "start", vb.options.StartKey)
	assert.Equal(t, "end", vb.options.EndKey)
	assert.Equal(t, 10, vb.options.Limit)
	assert.Equal(t, 5, vb.options.Skip)
	assert.True(t, vb.options.Descending)
	assert.True(t, vb.options.Group)
	assert.Equal(t, 2, vb.options.GroupLevel)
	assert.NotNil(t, vb.options.Reduce)
	assert.False(t, *vb.options.Reduce)
	assert.True(t, vb.options.IncludeDocs)
	assert.Equal(t, "ok", vb.options.Stale)

	// Test Keys method
	vb = db.NewViewQuery("test-design", "test-view").Keys("key1", "key2", "key3")
	assert.Equal(t, []interface{}{"key1", "key2", "key3"}, vb.options.Keys)
}

// Integration tests using the test suite
func (suite *CouchDBTestSuite) TestClient_Info() {
	info, err := suite.client.Info(context.Background())
	suite.Require().NoError(err)
	suite.NotEmpty(info.CouchDB)
	suite.NotEmpty(info.Version)
}

func (suite *CouchDBTestSuite) TestClient_AllDbs() {
	dbs, err := suite.client.AllDbs(context.Background())
	suite.Require().NoError(err)
	suite.NotEmpty(dbs)
}

func (suite *CouchDBTestSuite) TestDatabase_Lifecycle() {
	ctx := context.Background()

	// Create database
	err := suite.client.CreateDB(ctx, suite.testDB.name)
	if !suite.useMock {
		suite.Require().NoError(err)
	}

	// Get database info
	info, err := suite.testDB.Info(ctx)
	if !suite.useMock {
		suite.Require().NoError(err)
		suite.Equal(suite.testDB.name, info.DBName)
	}
}

func (suite *CouchDBTestSuite) TestDocument_CRUD() {
	if suite.useMock {
		suite.T().Skip("Skipping CRUD test with mock server")
	}

	ctx := context.Background()

	err := suite.client.CreateDB(ctx, suite.testDB.name)
	if err != nil {
		return
	}

	// Create document
	testDoc := map[string]interface{}{
		"name":  "Test Document",
		"type":  "test",
		"value": 42,
	}

	result, err := suite.testDB.Put(ctx, testDoc)
	suite.Require().NoError(err)
	suite.NotEmpty(result.ID)
	suite.NotEmpty(result.Rev)

	// Read document
	doc, err := suite.testDB.Get(ctx, result.ID)
	suite.Require().NoError(err)
	suite.Equal(result.ID, doc.ID)
	suite.Equal(result.Rev, doc.Rev)
	suite.Equal("Test Document", doc.Data["name"])

	// Update document
	updateDoc := map[string]interface{}{
		"_id":   result.ID,
		"_rev":  result.Rev,
		"name":  "Updated Document",
		"type":  "test",
		"value": 84,
	}

	updateResult, err := suite.testDB.Update(ctx, result.ID, updateDoc)
	suite.Require().NoError(err)
	suite.Equal(result.ID, updateResult.ID)
	suite.NotEqual(result.Rev, updateResult.Rev)

	// Delete document
	err = suite.testDB.Delete(ctx, updateResult.ID, updateResult.Rev)
	suite.Require().NoError(err)

	// Verify deletion
	_, err = suite.testDB.Get(ctx, result.ID)
	suite.Error(err)
}

func (suite *CouchDBTestSuite) TestDocument_Bulk() {
	if suite.useMock {
		suite.T().Skip("Skipping bulk test with mock server")
	}

	ctx := context.Background()
	err := suite.client.CreateDB(ctx, suite.testDB.name)
	if err != nil {
		return
	}

	docs := []interface{}{
		map[string]interface{}{
			"name": "Bulk Doc 1",
			"type": "bulk",
		},
		map[string]interface{}{
			"name": "Bulk Doc 2",
			"type": "bulk",
		},
		map[string]interface{}{
			"name": "Bulk Doc 3",
			"type": "bulk",
		},
	}

	results, err := suite.testDB.Bulk(ctx, docs)
	suite.Require().NoError(err)
	suite.Len(results, 3)

	for _, result := range results {
		suite.NotEmpty(result.ID)
		suite.NotEmpty(result.Rev)
		suite.Empty(result.Error)
	}
}

func (suite *CouchDBTestSuite) TestClient_UUID() {
	uuid, err := suite.client.UUID(context.Background())
	if !suite.useMock {
		suite.Require().NoError(err)
		suite.NotEmpty(uuid)
	}

	// Test multiple UUIDs
	uuids, err := suite.client.UUIDs(context.Background(), 3)
	if !suite.useMock {
		suite.Require().NoError(err)
		suite.Len(uuids, 3)
		for _, u := range uuids {
			suite.NotEmpty(u)
		}
	}
}

func (suite *CouchDBTestSuite) TestError_Handling() {
	ctx := context.Background()

	// Test getting non-existent document
	_, err := suite.testDB.Get(ctx, "non-existent-doc")
	suite.Error(err)

	var couchErr *Error
	if errors.As(err, &couchErr) {
		suite.Equal("not_found", couchErr.Type)
		suite.Contains(couchErr.Error(), "not_found")
	}
}

// Test Design Documents
func (suite *CouchDBTestSuite) TestDesignDocument() {
	if suite.useMock {
		suite.T().Skip("Skipping design document test with mock server")
	}

	ctx := context.Background()
	err := suite.client.CreateDB(ctx, suite.testDB.name)
	if err != nil {
		return
	}

	// Create a design document
	designDoc := &DesignDocument{
		Language: "javascript",
		Views: map[string]*View{
			"by_type": {
				Map:    "function(doc) { if (doc.type) emit(doc.type, 1); }",
				Reduce: "_count",
			},
		},
	}

	result, err := suite.testDB.PutDesignDoc(ctx, "test", designDoc)
	suite.Require().NoError(err)
	suite.NotEmpty(result.ID)
	suite.NotEmpty(result.Rev)

	// Get the design document
	retrievedDoc, err := suite.testDB.GetDesignDoc(ctx, "test")
	suite.Require().NoError(err)
	suite.Equal("_design/test", retrievedDoc.ID)
	suite.NotNil(retrievedDoc.Views["by_type"])

	// Test view query (after adding some test documents)
	testDocs := []interface{}{
		map[string]interface{}{"type": "user", "name": "John"},
		map[string]interface{}{"type": "user", "name": "Jane"},
		map[string]interface{}{"type": "post", "title": "Hello World"},
	}

	_, err = suite.testDB.Bulk(ctx, testDocs)
	suite.Require().NoError(err)

	// Query view
	viewResult, err := suite.testDB.View(ctx, "test", "by_type", &ViewOptions{
		Group: true,
	})
	suite.Require().NoError(err)
	suite.NotEmpty(viewResult.Rows)

	// Test convenience methods
	userResults, err := suite.testDB.ViewByKey(ctx, "test", "by_type", "user")
	suite.Require().NoError(err)
	suite.NotEmpty(userResults.Rows)

	// Test view builder
	builderResult, err := suite.testDB.NewViewQuery("test", "by_type").
		Key("user").
		IncludeDocs(true).
		Execute(ctx, suite.testDB)
	suite.Require().NoError(err)
	suite.NotEmpty(builderResult.Rows)
}

// Run the test suite
func TestCouchDBSuite(t *testing.T) {
	suite.Run(t, new(CouchDBTestSuite))
}

// Benchmark tests
func BenchmarkDocument_MarshalJSON(b *testing.B) {
	doc := &Document{
		ID:  "test-id",
		Rev: "1-abc123",
		Data: map[string]interface{}{
			"name":  "Test Document",
			"value": 42,
			"tags":  []string{"test", "benchmark"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := doc.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDocument_UnmarshalJSON(b *testing.B) {
	data := []byte(`{"_id":"test-id","_rev":"1-abc123","name":"Test Document","value":42,"tags":["test","benchmark"]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var doc Document
		err := doc.UnmarshalJSON(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Example test showing usage patterns
func ExampleClient_usage() {
	// Create client
	client := NewClient("http://localhost:5984", &ClientOptions{
		Username: "admin",
		Password: "password",
		Timeout:  30 * time.Second,
	})

	ctx := context.Background()

	// Get server info
	info, _ := client.Info(ctx)
	_ = info.Version

	// Work with database
	db := client.DB("my-database")

	// Create document
	doc := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}
	result, _ := db.Put(ctx, doc)
	_ = result.ID

	// Query view with builder pattern
	viewResult, _ := db.NewViewQuery("users", "by_age").
		StartKey(18).
		EndKey(65).
		IncludeDocs(true).
		Limit(10).
		Execute(ctx, db)

	_ = viewResult.Rows
}
