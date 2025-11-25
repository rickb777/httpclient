package jsoniter

import (
	xjsoniter "github.com/json-iterator/go"
)

// JSON is a configured jsoniter object.
//   - EscapeHTML: true
//   - CaseSensitive: true
var JSON = xjsoniter.Config{
	EscapeHTML:    true,
	CaseSensitive: true,
}.Froze()

// JSONIgnoreCase is a configured jsoniter object.
//   - EscapeHTML: true
//   - CaseSensitive: false
var JSONIgnoreCase = xjsoniter.Config{
	EscapeHTML:    true,
	CaseSensitive: false,
}.Froze()
