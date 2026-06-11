.PHONY: all build test test-orders test-products test-v coverage lint tidy proto clean

SERVICES := orders products

# ── build ──────────────────────────────────────────────────────────────────────

all: build

build:
	go build -C services/orders   -o ../../bin/orders   ./cmd/...
	go build -C services/products -o ../../bin/products ./cmd/...

# ── test ───────────────────────────────────────────────────────────────────────

test:
	go test -C services/orders   ./...
	go test -C services/products ./...

test-orders:
	go test -C services/orders ./...

test-products:
	go test -C services/products ./...

test-v:
	go test -C services/orders   -v ./...
	go test -C services/products -v ./...

coverage:
	go test -C services/orders   ./... -coverprofile=../../coverage-orders.out
	go test -C services/products ./... -coverprofile=../../coverage-products.out
	go tool cover -html=coverage-orders.out
	go tool cover -html=coverage-products.out

# ── lint ───────────────────────────────────────────────────────────────────────

lint:
	golangci-lint run ./services/orders/...
	golangci-lint run ./services/products/...

# ── tidy ───────────────────────────────────────────────────────────────────────

tidy:
	go mod tidy -C services/orders
	go mod tidy -C services/products
	go work sync

# ── proto ──────────────────────────────────────────────────────────────────────

PROTO_DIR := services/products/internal/grpc

proto:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/product.proto

# ── clean ──────────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ coverage-orders.out coverage-products.out
