# GO PROJECT

this is a learning Go project.

Info about tech stack and architecture can be found in [TECH.md](TECH.md)

## Basic Execution

Format code: `go fmt`

Run code: `go run .`

Build code: `go build`

Run exec: `./commerce-platform`

Run tests: `go test ./...`

### Products Service

Run with `go run ./services/products/cmd`

Run tests `go test ./services/products/...`

### Orders Service

Run with `go run ./services/orders/cmd`

Run tests `go test ./services/orders/...`

## GPRC

```bash
# install protoc
brew install protobuf
# install Go generators
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# add dependencies
go get google.golang.org/grpc
# create product.proto
# generate code
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  services/products/internal/grpc/product.proto
```
