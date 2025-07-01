package couchdb

import "context"

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
