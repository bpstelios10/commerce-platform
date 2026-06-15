.PHONY: all build run-all run-orders run-products test-all test-orders test-products test-shared test-v coverage lint tidy proto-server proto-client clean

SERVICES := orders products

# ── build ──────────────────────────────────────────────────────────────────────

all: build

build:
	go build -C services/orders   -o ../../bin/orders   ./cmd/...
	go build -C services/products -o ../../bin/products ./cmd/...

# ── run ────────────────────────────────────────────────────────────────────────

run-products:
	go run ./services/products/cmd/...

run-orders:
	go run ./services/orders/cmd/...

run-all:
	go run ./services/products/cmd/... &
	go run ./services/orders/cmd/...

# ── test ───────────────────────────────────────────────────────────────────────

test-all:
	go test -C shared            ./...
	go test -C services/orders   ./...
	go test -C services/products ./...

test-orders:
	go test -C services/orders ./...

test-products:
	go test -C services/products ./...

test-shared:
	go test -C shared ./...

test-v:
	go test -C shared            -v ./...
	go test -C services/orders   -v ./...
	go test -C services/products -v ./...

coverage:
	go test -C shared            ./... -coverprofile=../coverage-shared.out
	go test -C services/orders   ./... -coverprofile=../../coverage-orders.out
	go test -C services/products ./... -coverprofile=../../coverage-products.out
	go tool cover -html=coverage-shared.out
	go tool cover -html=coverage-orders.out
	go tool cover -html=coverage-products.out

# ── lint ───────────────────────────────────────────────────────────────────────

lint:
	golangci-lint run ./shared/...
	golangci-lint run ./services/orders/...
	golangci-lint run ./services/products/...

# ── tidy ───────────────────────────────────────────────────────────────────────

tidy:
	go mod tidy -C shared
	go mod tidy -C services/orders
	go mod tidy -C services/products
	go work sync

# ── proto ──────────────────────────────────────────────────────────────────────
# Usage:
#   make proto-server PROTO=product VERSION=v1
#   make proto-client PROTO=product VERSION=v1 SERVICE=orders

proto-server:
	protoc \
		--go_out=services/$(PROTO)s/internal/grpc \
		--go_opt=paths=source_relative \
		--go-grpc_out=services/$(PROTO)s/internal/grpc \
		--go-grpc_opt=paths=source_relative \
		--proto_path=protos/$(PROTO)/$(VERSION) \
		protos/$(PROTO)/$(VERSION)/$(PROTO).proto

proto-client:
	protoc \
		--go_out=services/$(SERVICE)/internal/grpc \
		--go_opt=paths=source_relative \
		--go-grpc_out=services/$(SERVICE)/internal/grpc \
		--go-grpc_opt=paths=source_relative \
		--proto_path=protos/$(PROTO)/$(VERSION) \
		protos/$(PROTO)/$(VERSION)/$(PROTO).proto

# ── clean ──────────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ coverage-shared.out coverage-orders.out coverage-products.out
