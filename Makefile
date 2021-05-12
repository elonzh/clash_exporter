.PHONY: lint snapshot
lint:
	golangci-lint run
snapshot:
	goreleaser --snapshot --rm-dist
