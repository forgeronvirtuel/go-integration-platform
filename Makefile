.PHONY: help build run clean test install dev

# Variables
BINARY_NAME=gip
GO=go
GOFLAGS=-v
BUILD_DIR=./bin
MAIN=main.go

help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compile le projet
	@echo "🔨 Compilation du projet..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN)
	@echo "✅ Binaire créé: $(BUILD_DIR)/$(BINARY_NAME)"

run: ## Lance le serveur en mode développement
	@echo "🚀 Démarrage du serveur..."
	CGO_ENABLED=1 $(GO) run $(MAIN) serve

dev: ## Lance le serveur avec rechargement automatique (nécessite air)
	@which air > /dev/null || (echo "⚠️  'air' n'est pas installé. Installez-le avec: go install github.com/air-verse/air@latest" && exit 1)
	air

install: ## Installe les dépendances
	@echo "📦 Installation des dépendances..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "✅ Dépendances installées"

test: ## Lance les tests unitaires
	@echo "🧪 Exécution des tests unitaires..."
	CGO_ENABLED=1 $(GO) test -v ./cmd/... ./internal/...

test-integration: ## Lance les tests d'intégration
	@echo "🧪 Exécution des tests d'intégration..."
	CGO_ENABLED=1 $(GO) test -v ./tests/integration/...

test-all: ## Lance tous les tests (unitaires + intégration)
	@echo "🧪 Exécution de tous les tests..."
	CGO_ENABLED=1 $(GO) test -v ./...

test-verbose: ## Lance les tests avec plus de détails
	@./test.sh

clean: ## Nettoie les fichiers générés
	@echo "🧹 Nettoyage..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f data.db
	@echo "✅ Nettoyage terminé"

lint: ## Vérifie le code avec golangci-lint
	@which golangci-lint > /dev/null || (echo "⚠️  'golangci-lint' n'est pas installé. Installez-le depuis: https://golangci-lint.run/welcome/install/" && exit 1)
	golangci-lint run

fmt: ## Formate le code
	@echo "✨ Formatage du code..."
	$(GO) fmt ./...
	@echo "✅ Code formaté"

vet: ## Analyse le code avec go vet
	@echo "🔍 Analyse du code..."
	$(GO) vet ./...
	@echo "✅ Analyse terminée"

check: fmt vet lint test ## Effectue toutes les vérifications (format, vet, lint, test)

all: clean install build ## Nettoie, installe et compile

.DEFAULT_GOAL := help
