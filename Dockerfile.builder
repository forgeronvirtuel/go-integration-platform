# syntax=docker/dockerfile:1

FROM golang:1.23-bookworm

# Create a non-root user for safety
RUN useradd -m -u 10001 -s /bin/bash builder

USER builder
WORKDIR /workspace

# Pre-create cache directories (optional but convenient)
ENV GOMODCACHE=/home/builder/.cache/gomod \
    GOCACHE=/home/builder/.cache/gocache

RUN mkdir -p "$GOMODCACHE" "$GOCACHE"

# bash command to build the docker image:
# DOCKER_BUILDKIT=1 docker build -f Dockerfile.builder -t builder-go:1.23 .