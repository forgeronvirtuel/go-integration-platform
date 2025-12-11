# Build API

Ce document explique comment utiliser l'API de build pour compiler des projets Go et télécharger les binaires générés.

## Endpoints

### 1. Créer un build

Déclenche la compilation d'un projet.

**Endpoint:** `POST /api/builds/`

**Request body:**

```json
{
  "project_id": 1
}
```

**Réponse (201 Created):**

```json
{
  "build_id": 1,
  "status": "success",
  "binary_path": "/workspace/project-1/out/mon-projet-1",
  "download_url": "/api/builds/1/download"
}
```

**Réponse en cas d'erreur (400 Bad Request):**

```json
{
  "error": "cmd/main.go not found in the repository",
  "logs": "==> Running: go mod download..."
}
```

### 2. Télécharger un binaire

Télécharge le binaire généré par un build réussi.

**Endpoint:** `GET /api/builds/:id/download`

**Paramètres:**

- `id`: ID du build

**Réponse (200 OK):**

- Télécharge le fichier binaire avec le nom `{project-name}-{build-id}`
- Headers:
  - `Content-Type: application/octet-stream`
  - `Content-Disposition: attachment; filename=mon-projet-1`

**Réponse en cas d'erreur:**

**404 Not Found:**

```json
{
  "error": "Build not found"
}
```

**400 Bad Request (build pas en succès):**

```json
{
  "error": "Build is not successful, cannot download binary"
}
```

**404 Not Found (binaire supprimé du disque):**

```json
{
  "error": "Binary file not found on disk"
}
```

## Workflow complet

### 1. Créer un projet

```bash
curl -X POST http://localhost:3000/api/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mon-api",
    "repo_url": "https://github.com/user/mon-api.git",
    "branch": "main"
  }'
```

### 2. Déclencher un build

```bash
curl -X POST http://localhost:3000/api/builds/ \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": 1
  }'
```

Réponse:

```json
{
  "build_id": 1,
  "status": "success",
  "binary_path": "/workspace/project-1/out/mon-api-1",
  "download_url": "/api/builds/1/download"
}
```

### 3. Télécharger le binaire

```bash
curl -O -J http://localhost:3000/api/builds/1/download
```

Cela télécharge le binaire avec le nom `mon-api-1`.

## Détails techniques

### Structure du workspace

Les builds sont organisés dans le workspace comme suit:

```
workspace/
├── project-1/
│   ├── .git/
│   ├── cmd/
│   │   └── main.go
│   ├── .gomodcache/
│   ├── .gocache/
│   └── out/
│       └── mon-api-1     # Binaire généré
├── project-2/
│   └── ...
```

### Processus de build

1. **Clonage**: Le repository Git est cloné dans `workspace/project-{id}`
2. **Validation**: Vérifie que `cmd/main.go` existe (ou `{subdir}/cmd/main.go` si subdir est défini)
3. **Téléchargement des modules**: Exécute `go mod download`
4. **Compilation**: Exécute `go build -o out/{project-name}-{build-id} ./cmd/main.go`
5. **Persistance**: Le chemin du binaire est stocké dans la base de données

### Variables d'environnement contrôlées

Pour la sécurité, le processus de build utilise un environnement minimal:

- `PATH`: Conservé du parent (pour trouver git/go)
- `HOME`: Défini au répertoire de travail
- `GOMODCACHE`: Cache des modules Go isolé par projet
- `GOCACHE`: Cache de compilation Go isolé par projet

### Timeout

Le build a un timeout de **5 minutes**. Si le build prend plus de temps, il sera annulé automatiquement.

### Stockage des logs

Les logs de compilation sont stockés dans le champ `log_output` de la base de données, avec le format:

```
Binary: /workspace/project-1/out/mon-api-1

==> Running: go mod download (in /workspace/project-1)
...
==> Running: go build ... (in /workspace/project-1)
...
```

## Prérequis pour les projets

Pour qu'un projet puisse être compilé, il doit:

1. Avoir un fichier `cmd/main.go` (ou `{subdir}/cmd/main.go` si un sous-répertoire est configuré)
2. Avoir un fichier `go.mod` valide
3. Être compilable avec `go build`

## Exemples de gestion d'erreurs

### Repository inaccessible

```json
{
  "error": "Failed to clone repository"
}
```

### Fichier main.go manquant

```json
{
  "error": "cmd/main.go not found in the repository"
}
```

### Erreur de compilation

```json
{
  "error": "exit status 1",
  "logs": "==> Running: go build ...\n# mon-api\n./main.go:10:2: undefined: SomeFunction\n"
}
```
