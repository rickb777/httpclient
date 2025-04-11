package logging

// see github.com/rickb777/enumeration/v2
//go:generate enumeration -v -type Level -ic -s

// Level allows control of the predicate of detail in log messages.
type Level int

const (
	// Off turns logging off.
	Off Level = iota

	// Discrete log messages contain only a summary of the request and response.
	// No query parameters are printed in order to hide potential personal information.
	Discrete

	// Summary log messages contain only a summary of the request and response,
	// including the full target URL.
	Summary

	// WithHeaders log messages contain a summary and the request/response headers
	WithHeaders

	// WithHeadersAndBodies log messages contain a summary and the request/response headers and bodies
	// Textual bodies are included in the log; for binary content, the size is shown instead.
	WithHeadersAndBodies
)
