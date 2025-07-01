package couchdb

import "context"

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

// ListDesignocs lists all design documents
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
