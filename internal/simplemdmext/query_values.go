package simplemdmext

import "net/url"

// queryValues is a thin wrapper around url.Values that adds typed helpers.
type queryValues struct {
	url.Values
}

func newQueryValues() queryValues {
	return queryValues{Values: url.Values{}}
}

// SetBool sets a query parameter using the canonical "true"/"false" strings the
// SimpleMDM API expects for boolean fields.
func (q queryValues) SetBool(key string, value bool) {
	if value {
		q.Set(key, "true")
		return
	}
	q.Set(key, "false")
}
