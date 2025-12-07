#!/bin/bash

# Script de build pour Go Integration Platform
set -e

BINARY_NAME="gip"
BUILD_DIR="./bin"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | awk '{print $3}')

echo "🔨 Building Go Integration Platform"
echo "   Version: ${VERSION}"
echo "   Build Time: ${BUILD_TIME}"
echo "   Go Version: ${GO_VERSION}"
echo ""

# Créer le répertoire de build
mkdir -p ${BUILD_DIR}

# Flags de build
LDFLAGS="-s -w"
LDFLAGS="${LDFLAGS} -X main.Version=${VERSION}"
LDFLAGS="${LDFLAGS} -X main.BuildTime=${BUILD_TIME}"

# Build pour la plateforme actuelle
echo "📦 Building for current platform..."
CGO_ENABLED=1 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME} .

# Build pour d'autres plateformes (optionnel)
if [ "$1" == "all" ]; then
    echo ""
    echo "📦 Building for multiple platforms..."
    
    # Linux AMD64
    echo "  → Linux AMD64"
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 .
    
    # Linux ARM64
    echo "  → Linux ARM64"
    CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-linux-arm64 .
    
    # macOS AMD64
    echo "  → macOS AMD64"
    CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 .
    
    # macOS ARM64 (Apple Silicon)
    echo "  → macOS ARM64"
    CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64 .
    
    # Windows AMD64
    echo "  → Windows AMD64"
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe .
fi

echo ""
echo "✅ Build complete!"
echo "   Binary location: ${BUILD_DIR}/${BINARY_NAME}"
echo ""
echo "Run with: ${BUILD_DIR}/${BINARY_NAME} serve"
