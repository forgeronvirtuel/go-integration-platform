# Go Integration Platform (GIP)

Un serveur HTTP minimaliste construit avec Go, utilisant :

- **Gin** : Framework web léger et performant
- **Cobra** : Gestion de la ligne de commande
- **Zerolog** : Logging structuré et performant
- **SQLite** : Base de données embarquée

## Installation

```bash
# Installer les dépendances
make install
# ou
go mod download
```

## Build

### Avec Make (recommandé)

```bash
# Voir toutes les commandes disponibles
make help

# Compiler le projet
make build

# Compiler et installer
make all
```

### Avec le script build.sh

```bash
# Build pour la plateforme actuelle
./build.sh

# Build pour toutes les plateformes (Linux, macOS, Windows)
./build.sh all
```

### Build manuel

```bash
go build -o bin/gip main.go
```

### Avec Docker

```bash
# Construire l'image
docker build -t gip:latest .

# Lancer le conteneur
docker run -p 8080:8080 -v $(pwd)/data:/data gip:latest
```

## Utilisation

### Démarrer le serveur

```bash
# Avec Make
make run

# Avec le binaire compilé
./bin/gip serve

# En mode développement (go run)
go run main.go serve
```

Options disponibles :

- `-p, --port` : Port d'écoute (défaut: 3000)
- `-d, --database` : Chemin vers le fichier SQLite (défaut: ./data.db)
- `-w, --workspace` : Répertoire de workspace pour les projets (défaut: ./workspace)

Exemple :

```bash
./bin/gip serve --port 3000 --database /tmp/mydb.db --workspace /var/projects
```

**Note importante sur le workspace :**

Le programme vérifie au démarrage que le répertoire de workspace existe et dispose des permissions en lecture/écriture. Si le répertoire n'existe pas, il sera créé automatiquement. Si le programme n'a pas les permissions nécessaires, il s'arrêtera avec une erreur.

### Développement avec rechargement automatique

```bash
# Installer Air (si pas déjà fait)
go install github.com/air-verse/air@latest

# Lancer en mode dev
make dev
```

## Tests

### Lancer les tests

```bash
# Tests simples
make test

# Tests avec détails
make test-verbose
# ou
./test.sh

# Couverture de code
make coverage
# ou
./coverage.sh
```

### Structure des tests

Les tests sont organisés par package :

- `cmd/` : Tests des commandes CLI
- `internal/database/` : Tests de la couche base de données
- `internal/server/` : Tests des routes HTTP

Tous les tests utilisent :

- **testify** pour les assertions
- **SQLite en mémoire** pour les tests de DB
- **httptest** pour les tests HTTP

## Endpoints disponibles

- `GET /` : Hello World
- `GET /health` : Vérification de l'état du serveur et de la DB
- `GET /users` : Liste des utilisateurs (données de test)

## Commandes Make disponibles

```bash
make help            # Affiche l'aide
make build           # Compile le projet
make run             # Lance le serveur en mode développement
make dev             # Lance avec rechargement automatique (nécessite air)
make install         # Installe les dépendances
make test            # Lance les tests unitaires
make test-integration # Lance les tests d'intégration HTTP
make test-all        # Lance tous les tests
make test-verbose    # Lance les tests avec plus de détails
make coverage        # Génère un rapport de couverture de code
make clean           # Nettoie les fichiers générés
make fmt             # Formate le code
make vet             # Analyse le code
make lint            # Vérifie le code (nécessite golangci-lint)
make check           # Effectue toutes les vérifications
make all             # Nettoie, installe et compile
```

## Structure du projet

```
.
├── cmd/
│   ├── root.go          # Commande racine Cobra
│   └── serve.go         # Commande pour démarrer le serveur
├── internal/
│   ├── database/
│   │   └── database.go  # Initialisation SQLite et gestion des tables
│   └── server/
│       └── server.go    # Configuration du serveur Gin
├── main.go              # Point d'entrée
├── Makefile             # Commandes de build
├── build.sh             # Script de build multi-plateforme
├── Dockerfile           # Configuration Docker
├── .air.toml            # Configuration Air (rechargement auto)
└── go.mod
```

## Base de données

La base de données SQLite est automatiquement initialisée au démarrage avec une table `users` contenant des données de test.

Structure de la table :

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```
