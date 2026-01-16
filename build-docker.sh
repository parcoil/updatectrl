#!/bin/bash

# Build multi-platform Docker image for Updatectrl

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag ghcr.io/parcoil/updatectrl:latest \
  --push \
  .