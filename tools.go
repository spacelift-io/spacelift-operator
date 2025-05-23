//go:build tools

package tools

//go:generate go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
//go:generate go install github.com/daixiang0/gci@v0.13.6
//go:generate go install github.com/goreleaser/goreleaser/v2@v2.9.0
//go:generate go install github.com/vektra/mockery/v2@v2.53.4
