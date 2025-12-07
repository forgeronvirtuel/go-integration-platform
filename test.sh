#!/bin/bash

# Script de test pour Go Integration Platform
set -e

echo "🧪 Running tests for Go Integration Platform"
echo ""

# Couleurs
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Fonction pour exécuter les tests d'un package
run_tests() {
    local package=$1
    echo -e "${YELLOW}Testing ${package}...${NC}"
    if CGO_ENABLED=1 go test -v "$package"; then
        echo -e "${GREEN}✅ ${package} tests passed${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}❌ ${package} tests failed${NC}"
        echo ""
        return 1
    fi
}

# Exécuter les tests par package
FAILED=0

run_tests "./cmd/..." || FAILED=1
run_tests "./internal/database/..." || FAILED=1
run_tests "./internal/server/..." || FAILED=1

# Résumé
echo "═══════════════════════════════════════"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
