//go:build tools

package tools

//go:generate go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
//go:generate go install github.com/daixiang0/gci@v0.11.2
//go:generate go install github.com/goreleaser/goreleaser@v1.24.0
//go:generate go install github.com/vektra/mockery/v2@v2.40.2
