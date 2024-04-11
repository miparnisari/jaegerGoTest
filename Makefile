.PHONY: buf-generate
buf-generate: ## Generate source files from protobuf sources
	@buf generate

.PHONY: build-protos
build-protos: buf-generate ## Build/generate source files from protobuf sources

.PHONY: build
build: build-protos ## Build
	go build -o ./server .