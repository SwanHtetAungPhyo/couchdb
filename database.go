package couchdb

import "context"

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
