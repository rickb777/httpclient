module github.com/rickb777/httpclient

go 1.24.1

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/go-xmlfmt/xmlfmt v1.1.3
	github.com/magefile/mage v1.15.0
	github.com/rickb777/expect v0.24.0
	github.com/rs/zerolog v1.34.0
	github.com/spf13/afero v1.14.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rickb777/enumeration v1.10.2 // indirect
	github.com/rickb777/enumeration/v4 v4.0.4 // indirect
	github.com/rickb777/plural v1.4.4 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)

tool (
	github.com/magefile/mage
	github.com/rickb777/enumeration/v4
)
