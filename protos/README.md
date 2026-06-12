# PROTOS MODULE

This is a folder where the shared protos live.

Standard pattern for a mono-repo project.

## COMMANDS

- generate server code into services/products/internal/grpc/
  
  `make proto-server PROTO=product VERSION=v1`

- generate client code into services/orders/internal/grpc/
  
  `make proto-client PROTO=product VERSION=v1 SERVICE=orders`

- later when v2 ships, products can run both:
  
  `make proto-server PROTO=product VERSION=v2`

- orders migrates independently whenever ready:
  
  `make proto-client PROTO=product VERSION=v2 SERVICE=orders`
