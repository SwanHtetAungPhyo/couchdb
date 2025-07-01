# CouchDB Go Client

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/couchdb-go)](https://goreportcard.com/report/github.com/yourusername/couchdb-go)
[![GoDoc](https://godoc.org/github.com/yourusername/couchdb-go?status.svg)](https://godoc.org/github.com/yourusername/couchdb-go)

A comprehensive and intuitive Go client library for Apache CouchDB, built on top of the Resty HTTP client. This library provides a clean, fluent API for all CouchDB operations including documents, views, design documents, and database administration.

## ✨ Features

- 🚀 **Full CouchDB API Coverage** - Complete support for all CouchDB operations
- 🏗️ **Fluent Query Builder** - Intuitive builder pattern for complex view queries
- 📦 **Bulk Operations** - Efficient bulk document operations
- 🎯 **Context Support** - Full context.Context support for all operations
- 🔍 **Advanced View Queries** - Comprehensive view query options and helpers
- 📋 **Design Document Management** - Full CRUD operations for design documents
- ⚡ **Built-in Error Handling** - Structured error responses with CouchDB error details
- 🛠️ **Database Administration** - Database creation, deletion, compaction, and maintenance

## 🚀 Quick Start

### Installation

```bash
go get github.com/SwanHtetAungPhyo/couchdb
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/SwanHtetAungPhyo/couchdb"
)

func main() {
    // Create client
    client := couchdb.NewClient("http://localhost:5984", &couchdb.ClientOptions{
        Username: "admin",
        Password: "password",
        Timeout:  30 * time.Second,
    })

    // Get database reference
    db := client.DB("myapp")
    ctx := context.Background()

    // Create a document
    user := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
    }

    result, err := db.Put(ctx, user)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created user: %s\n", result.ID)

    // Retrieve the document
    doc, err := db.Get(ctx, result.ID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("User: %+v\n", doc.Data)
}
```

## 📖 Documentation

### Client Configuration

```go
client := couchdb.NewClient("http://localhost:5984", &couchdb.ClientOptions{
    Username: "admin",           // Basic auth username
    Password: "password",        // Basic auth password
    Timeout:  30 * time.Second, // Request timeout
    Debug:    false,            // Enable debug logging
})
```

### Document Operations

```go
db := client.DB("mydb")
ctx := context.Background()

// Create document
doc := map[string]interface{}{"name": "Alice", "age": 25}
result, err := db.Put(ctx, doc)

// Get document
doc, err := db.Get(ctx, "doc-id")

// Update document
updated := map[string]interface{}{
    "_id": doc.ID,
    "_rev": doc.Rev,
    "name": "Alice Smith",
    "age": 26,
}
result, err := db.Update(ctx, doc.ID, updated)

// Delete document
err = db.Delete(ctx, doc.ID, doc.Rev)

// Bulk operations
docs := []interface{}{
    map[string]interface{}{"name": "Bob"},
    map[string]interface{}{"name": "Carol"},
}
results, err := db.Bulk(ctx, docs)
```

### View Queries

#### Simple View Queries

```go
// Get all results
result, err := db.ViewAll(ctx, "users", "by_age", true)

// Query by key
result, err := db.ViewByKey(ctx, "users", "by_city", "New York")

// Query by range
result, err := db.ViewByKeyRange(ctx, "users", "by_age", 25, 35)

// Get reduced results
result, err := db.ViewReduce(ctx, "stats", "count_by_city", 1)
```

#### Advanced Query Builder

```go
result, err := db.NewViewQuery("users", "by_age").
    StartKey(25).
    EndKey(40).
    IncludeDocs(true).
    Limit(10).
    Descending(true).
    Execute(ctx, db)

// Complex query with multiple options
result, err := db.NewViewQuery("products", "by_category").
    Key("electronics").
    IncludeDocs(true).
    Stale("update_after").
    Group(true).
    Execute(ctx, db)
```

#### Direct View Query with Options

```go
opts := &couchdb.ViewOptions{
    StartKey:    "A",
    EndKey:      "M",
    IncludeDocs: true,
    Limit:       50,
    Descending:  false,
}

result, err := db.View(ctx, "mydesign", "myview", opts)
```

### Design Documents

```go
// Create design document with views
designDoc := &couchdb.DesignDocument{
    Language: "javascript",
    Views: map[string]*couchdb.View{
        "by_name": {
            Map: `function(doc) {
                if (doc.name) {
                    emit(doc.name, doc);
                }
            }`,
        },
        "count_by_status": {
            Map: `function(doc) {
                if (doc.status) {
                    emit(doc.status, 1);
                }
            }`,
            Reduce: "_count",
        },
    },
}

// Save design document
result, err := db.PutDesignDoc(ctx, "users", designDoc)

// Get design document
designDoc, err := db.GetDesignDoc(ctx, "users")

// List all design documents
designs, err := db.ListDesignDocs(ctx)

// Delete design document
err = db.DeleteDesignDoc(ctx, "users", rev)
```

### Database Administration

```go
// Server info
info, err := client.Info(ctx)
fmt.Printf("CouchDB %s\n", info.Version)

// List databases
dbs, err := client.AllDbs(ctx)

// Create/delete database
err = client.CreateDB(ctx, "newdb")
err = client.DeleteDB(ctx, "olddb")

// Database info
dbInfo, err := db.Info(ctx)
fmt.Printf("Documents: %d\n", dbInfo.DocCount)

// Maintenance operations
err = db.Compact(ctx)                    // Compact database
err = db.CompactDesignDoc(ctx, "users")  // Compact design doc
err = db.ViewCleanup(ctx)                // Clean up old view files
```

### Changes Feed

```go
changes, err := db.Changes(ctx, map[string]interface{}{
    "since":        "now",
    "feed":         "normal",
    "include_docs": true,
    "limit":        100,
})
```

### Error Handling

```go
doc, err := db.Get(ctx, "nonexistent")
if err != nil {
    if couchErr, ok := err.(*couchdb.Error); ok {
        switch couchErr.StatusCode {
        case 404:
            fmt.Println("Document not found")
        case 409:
            fmt.Println("Document conflict")
        default:
            fmt.Printf("CouchDB error: %s - %s\n", couchErr.Type, couchErr.Reason)
        }
    }
}
```

## 🔧 Complete Examples

<details>
<summary><strong>📝 Blog Application Example</strong></summary>

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/SwanHtetAungPhy"
)

type BlogPost struct {
    ID        string    `json:"_id,omitempty"`
    Rev       string    `json:"_rev,omitempty"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Author    string    `json:"author"`
    Tags      []string  `json:"tags"`
    Published bool      `json:"published"`
    CreatedAt time.Time `json:"created_at"`
}

func main() {
    client := couchdb.NewClient("http://localhost:5984", &couchdb.ClientOptions{
        Username: "admin",
        Password: "password",
    })

    db := client.DB("blog")
    ctx := context.Background()

    // Create database
    err := client.CreateDB(ctx, "blog")
    if err != nil {
        log.Printf("Database might already exist: %v", err)
    }

    // Create design document for blog queries
    designDoc := &couchdb.DesignDocument{
        Views: map[string]*couchdb.View{
            "by_author": {
                Map: `function(doc) {
                    if (doc.author && doc.published) {
                        emit(doc.author, {
                            title: doc.title,
                            created_at: doc.created_at
                        });
                    }
                }`,
            },
            "by_tag": {
                Map: `function(doc) {
                    if (doc.tags && doc.published) {
                        doc.tags.forEach(function(tag) {
                            emit(tag, {
                                title: doc.title,
                                author: doc.author
                            });
                        });
                    }
                }`,
            },
            "published_count": {
                Map: `function(doc) {
                    if (doc.published !== undefined) {
                        emit(doc.published, 1);
                    }
                }`,
                Reduce: "_count",
            },
        },
    }

    _, err = db.PutDesignDoc(ctx, "blog", designDoc)
    if err != nil {
        log.Fatal(err)
    }

    // Create blog posts
    posts := []BlogPost{
        {
            Title:     "Getting Started with CouchDB",
            Content:   "CouchDB is a NoSQL database...",
            Author:    "alice",
            Tags:      []string{"couchdb", "nosql", "database"},
            Published: true,
            CreatedAt: time.Now(),
        },
        {
            Title:     "Advanced CouchDB Views",
            Content:   "Views in CouchDB are powerful...",
            Author:    "bob",
            Tags:      []string{"couchdb", "views", "advanced"},
            Published: true,
            CreatedAt: time.Now(),
        },
        {
            Title:     "Draft Post",
            Content:   "This is a draft...",
            Author:    "alice",
            Tags:      []string{"draft"},
            Published: false,
            CreatedAt: time.Now(),
        },
    }

    // Bulk insert posts
    docs := make([]interface{}, len(posts))
    for i, post := range posts {
        docs[i] = post
    }

    results, err := db.Bulk(ctx, docs)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created %d blog posts\n", len(results))

    // Query posts by author
    authorPosts, err := db.ViewByKey(ctx, "blog", "by_author", "alice")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Alice has %d published posts\n", len(authorPosts.Rows))

    // Query posts by tag with limit
    tagPosts, err := db.NewViewQuery("blog", "by_tag").
        Key("couchdb").
        Limit(5).
        Execute(ctx, db)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d posts tagged 'couchdb'\n", len(tagPosts.Rows))

    // Get publication statistics
    stats, err := db.ViewReduce(ctx, "blog", "published_count", 0)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Publication stats:")
    for _, row := range stats.Rows {
        status := "draft"
        if row.Key.(bool) {
            status = "published"
        }
        fmt.Printf("  %s: %.0f posts\n", status, row.Value.(float64))
    }
}
```

</details>

<details>
<summary><strong>🛒 E-commerce Product Catalog Example</strong></summary>

```go
package main

import (
	"context"
	"fmt"
	"github.com/SwanHtetAungPhyo/couchdb/couchdb"
	"log"

	"github.com/yourusername/couchdb-go"
)

type Product struct {
	ID          string  `json:"_id,omitempty"`
	Rev         string  `json:"_rev,omitempty"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	InStock     bool    `json:"in_stock"`
	Description string  `json:"description"`
	Brand       string  `json:"brand"`
}

func main() {
	client := couchdb.NewClient("http://localhost:5984", nil)
	db := client.DB("products")
	ctx := context.Background()

	// Create products design document
	designDoc := &couchdb.DesignDocument{
		Views: map[string]*couchdb.View{
			"by_category": {
				Map: `function(doc) {
                    if (doc.category && doc.in_stock) {
                        emit([doc.category, doc.price], {
                            name: doc.name,
                            brand: doc.brand,
                            price: doc.price
                        });
                    }
                }`,
			},
			"by_price_range": {
				Map: `function(doc) {
                    if (doc.price && doc.in_stock) {
                        emit(doc.price, {
                            name: doc.name,
                            category: doc.category,
                            brand: doc.brand
                        });
                    }
                }`,
			},
			"total_value_by_category": {
				Map: `function(doc) {
                    if (doc.category && doc.price && doc.in_stock) {
                        emit(doc.category, doc.price);
                    }
                }`,
				Reduce: "_sum",
			},
		},
	}

	_, err := db.PutDesignDoc(ctx, "products", designDoc)
	if err != nil {
		log.Fatal(err)
	}

	// Sample products
	products := []Product{
		{Name: "iPhone 14", Category: "electronics", Price: 999.99, InStock: true, Brand: "Apple"},
		{Name: "MacBook Pro", Category: "electronics", Price: 2399.99, InStock: true, Brand: "Apple"},
		{Name: "Samsung TV", Category: "electronics", Price: 1299.99, InStock: false, Brand: "Samsung"},
		{Name: "Coffee Maker", Category: "appliances", Price: 79.99, InStock: true, Brand: "Breville"},
		{Name: "Blender", Category: "appliances", Price: 129.99, InStock: true, Brand: "Vitamix"},
	}

	// Insert products
	docs := make([]interface{}, len(products))
	for i, product := range products {
		docs[i] = product
	}

	_, err = db.Bulk(ctx, docs)
	if err != nil {
		log.Fatal(err)
	}

	// Query electronics under $2000
	electronics, err := db.NewViewQuery("products", "by_category").
		StartKey([]interface{}{"electronics", 0}).
		EndKey([]interface{}{"electronics", 2000}).
		IncludeDocs(false).
		Execute(ctx, db)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Electronics under $2000:\n")
	for _, row := range electronics.Rows {
		product := row.Value.(map[string]interface{})
		fmt.Printf("  %s - $%.2f (%s)\n",
			product["name"], product["price"], product["brand"])
	}

	// Get products in price range $100-$500
	priceRange, err := db.NewViewQuery("products", "by_price_range").
		StartKey(100).
		EndKey(500).
		Execute(ctx, db)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nProducts $100-$500:\n")
	for _, row := range priceRange.Rows {
		product := row.Value.(map[string]interface{})
		fmt.Printf("  %s - $%.2f (%s)\n",
			product["name"], row.Key, product["category"])
	}

	// Get total inventory value by category
	categoryTotals, err := db.ViewReduce(ctx, "products", "total_value_by_category", 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nInventory value by category:\n")
	for _, row := range categoryTotals.Rows {
		fmt.Printf("  %s: $%.2f\n", row.Key, row.Value)
	}
}
```

</details>

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙋‍♂️ Support

- 📖 [Documentation](https://pkg.go.dev/github.com/yourusername/couchdb-go)
- 🐛 [Issue Tracker](https://github.com/yourusername/couchdb-go/issues)
- 💬 [Discussions](https://github.com/yourusername/couchdb-go/discussions)

## 🔗 Related Projects

- [Apache CouchDB](https://couchdb.apache.org/) - The database this client connects to
- [go-resty/resty](https://github.com/go-resty/resty) - HTTP client library used internally

---

<p align="center">Made with ❤️ for the Go and CouchDB communities</p>