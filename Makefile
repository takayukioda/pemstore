.PHONY: fmt lint

PACKAGES = ./...

lint:
	go vet $(PACKAGES)
fmt: lint
	go fmt $(PACKAGES)
