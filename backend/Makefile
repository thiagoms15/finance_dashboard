APP_NAME=finance-backend

.PHONY: run-api run-worker migrate-up test fmt vet

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

migrate-up:
	go run ./cmd/migrate

test:
	go test ./...

fmt:
	gofmt -w $$(rg --files -g '*.go')

vet:
	go vet ./...
