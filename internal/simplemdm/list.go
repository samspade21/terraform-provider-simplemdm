package simplemdm

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DecodeList parses a SimpleMDM list response into a slice of T.
//
// SimpleMDM list endpoints normally return an envelope:
//
//	{"data": [...], "has_more": <bool>}
//
// but at least the /custom_attributes endpoint returns a bare array
// instead, despite what the API spec says. This helper accepts either
// shape: it peeks at the first non-whitespace byte to decide whether it's
// looking at a `[` (bare array) or anything else (envelope).
//
// Returns (items, hasMore, err). When the response is a bare array,
// hasMore is always false (bare-array responses do not paginate).
func DecodeList[T any](body []byte) ([]T, bool, error) {
	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var items []T
		if err := json.Unmarshal(body, &items); err != nil {
			return nil, false, fmt.Errorf("decode bare-array list response: %w", err)
		}
		return items, false, nil
	}

	var envelope struct {
		Data    []T  `json:"data"`
		HasMore bool `json:"has_more"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, false, fmt.Errorf("decode list envelope: %w", err)
	}
	return envelope.Data, envelope.HasMore, nil
}
