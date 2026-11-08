//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --parseInternal --requiredByDefault -g internal/route/route.go --output docs --outputTypes json,yaml --exclude web,tools

package main
