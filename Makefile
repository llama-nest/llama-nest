APP=llama-nest

.PHONY: fmt
fmt:
	gofmt -w ./cmd ./internal

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: build
build: fmt vet test
	go build -o bin/$(APP) ./cmd/llama-nest

.PHONY: run
run:
	go run ./cmd/llama-nest start

.PHONY: clean
clean:
	rm -rf bin

.PHONY: transfer-test
transfer-test:
	curl http://localhost:11435/api/chat \
		-d '{"model":"llama3.2","messages":[{"role":"user","content":"test transfer"}],"stream":false}'

	go run ./cmd/llama-nest transfer qwen2.5:7b
