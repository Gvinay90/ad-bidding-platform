#!/bin/bash
set -e
for svc in campaign bidder analytics; do
  protoc \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/$svc/$svc.proto
done
