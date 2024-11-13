//go:build tools

package tools

//go:generate go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.0
//go:generate go install github.com/daixiang0/gci@v0.13.5
//go:generate go install github.com/goreleaser/goreleaser/v2@v2.4.4
//go:generate go install github.com/vektra/mockery/v2@v2.46.3
