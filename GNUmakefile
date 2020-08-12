include ./opinionated.mk

test:
> go test -timeout 30s ./...
.PHONY: test

fmt:
> go fmt ./pkg/...
.PHONY: fmt

lint:
> golint ./pkg/...
.PHONY: lint

coverage:
> go test -timeout 30s -coverprofile /dev/null ./pkg/...
.PHONY: coverage

viewcoverage:
> go test -timeout 30s -coverprofile cover.out ./pkg/...
> go tool cover -html=cover.out
.PHONY: viewcoverage

