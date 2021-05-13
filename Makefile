.PHONY: lint test snapshot
lint:
	golangci-lint run

test:
	go test -coverprofile=coverage.txt -covermode=atomic

snapshot:
	goreleaser --snapshot --rm-dist
