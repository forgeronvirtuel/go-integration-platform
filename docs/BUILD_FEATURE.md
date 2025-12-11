# Fonctionnalité de Build et Téléchargement

## Vue d'ensemble

La fonction `CreateBuild` permet de :

1. Cloner un repository Git
2. Compiler un binaire Go à partir du code source
3. Stocker le binaire généré
4. Permettre le téléchargement du binaire via une API REST

## Architecture

### Composants créés

1. **`internal/database/builds.go`**

   - Structure `Build` pour représenter un build en base de données
   - Fonctions CRUD : `CreateBuild`, `GetBuildByID`, `UpdateBuildStatus`, etc.
   - Gestion de la table `builds` avec foreign key vers `projects`

2. **`internal/server/builds.go`** (mis à jour)

   - `CreateBuild()` : Clone le repo, compile le binaire, stocke le résultat
   - `DownloadBinary()` : Permet de télécharger le binaire généré
   - `setupBuildRoutes()` : Configure les routes `/api/builds/` et `/api/builds/:id/download`

3. **Documentation**
   - `docs/BUILD_API.md` : Documentation complète de l'API
   - Exemples d'utilisation avec curl

### Flux de données

```
POST /api/builds/
  → CreateBuild()
    → Clone repository git
    → Vérifie cmd/main.go
    → go mod download
    → go build
    → Stocke le chemin du binaire en DB
  → Retourne build_id + download_url

GET /api/builds/:id/download
  → DownloadBinary()
    → Récupère le build depuis DB
    → Vérifie le statut (success)
    → Extrait le chemin du binaire
    → Envoie le fichier
```

## Sécurité et isolation

### Environnement contrôlé

Le build s'exécute avec un environnement minimal :

- `PATH` : Conservé du parent (pour trouver `git` et `go`)
- `HOME` : Défini au répertoire de travail
- `GOMODCACHE` : Cache des modules isolé par projet
- `GOCACHE` : Cache de compilation isolé

### Timeout

- Build timeout : **5 minutes**
- Évite les processus infinis

### Validation

- Vérifie l'existence de `cmd/main.go`
- Vérifie les permissions sur le workspace
- Pas d'exécution de shell (`bash -c`) pour éviter les injections

## Structure du workspace

```
workspace/
├── project-1/
│   ├── .git/              # Repository cloné
│   ├── cmd/
│   │   └── main.go
│   ├── .gomodcache/       # Cache modules Go isolé
│   ├── .gocache/          # Cache compilation isolé
│   └── out/
│       └── project-1-1    # Binaire généré (format: {project-name}-{build-id})
├── project-2/
│   └── ...
```

## Base de données

### Table `builds`

```sql
CREATE TABLE builds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    branch TEXT NOT NULL,
    status TEXT DEFAULT 'pending',    -- pending, building, success, failed
    log_output TEXT,                  -- Logs de compilation + chemin du binaire
    started_at DATETIME,
    ended_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

### Statuts

- `pending` : Build créé, pas encore démarré
- `building` : Compilation en cours
- `success` : Build réussi, binaire disponible
- `failed` : Erreur lors de la compilation

### Format de `log_output`

Pour un build réussi :

```
Binary: /workspace/project-1/out/project-1-1

==> Running: go mod download (in /workspace/project-1)
...
==> Running: go build ... (in /workspace/project-1)
...
```

## API Endpoints

### 1. POST `/api/builds/`

Crée et exécute un build.

**Request:**

```json
{
  "project_id": 1
}
```

**Response (201):**

```json
{
  "build_id": 1,
  "status": "success",
  "binary_path": "/workspace/project-1/out/mon-projet-1",
  "download_url": "/api/builds/1/download"
}
```

**Response (400) - Erreur:**

```json
{
  "error": "cmd/main.go not found in the repository",
  "logs": "..."
}
```

### 2. GET `/api/builds/:id/download`

Télécharge le binaire généré.

**Response (200):**

- Headers :
  - `Content-Type: application/octet-stream`
  - `Content-Disposition: attachment; filename={project-name}-{build-id}`
- Body : Fichier binaire

**Erreurs possibles:**

- `404` : Build non trouvé
- `400` : Build pas en statut "success"
- `404` : Fichier binaire introuvable sur le disque

## Tests

### Tests d'intégration

Fichier : `tests/integration/build_test.go`

Tests implémentés :

1. `TestIntegrationBuildAndDownload` : Workflow complet création projet → build → téléchargement
2. `TestBuildEndpoint_ProjectNotFound` : Gestion d'erreur projet inexistant
3. `TestDownloadBinary_BuildNotFound` : Gestion d'erreur build inexistant

## Utilisation

### Exemple complet

```bash
# 1. Démarrer le serveur
./bin/gip serve --workspace /var/builds --port 3000

# 2. Créer un projet
curl -X POST http://localhost:3000/api/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mon-api",
    "repo_url": "https://github.com/user/mon-api.git",
    "branch": "main"
  }'

# Response: {"id": 1, ...}

# 3. Déclencher un build
curl -X POST http://localhost:3000/api/builds/ \
  -H "Content-Type: application/json" \
  -d '{"project_id": 1}'

# Response:
# {
#   "build_id": 1,
#   "status": "success",
#   "download_url": "/api/builds/1/download"
# }

# 4. Télécharger le binaire
curl -O -J http://localhost:3000/api/builds/1/download

# Fichier téléchargé: mon-api-1
```

## Prérequis

### Pour le serveur

- Go installé (pour `go build`)
- Git installé (pour `git clone`)
- Permissions lecture/écriture sur le workspace

### Pour les projets à compiler

- Fichier `cmd/main.go` présent
- Fichier `go.mod` valide
- Code compilable avec `go build`

## Améliorations possibles

1. **Builds asynchrones** : Actuellement le build bloque la requête HTTP
   - Solution : Queue de builds avec workers
2. **Nettoyage automatique** : Les binaires s'accumulent dans le workspace
   - Solution : Cron job pour supprimer les vieux builds
3. **Support multi-plateforme** : Compiler pour Linux, macOS, Windows
   - Solution : Utiliser `GOOS` et `GOARCH` dans le build
4. **Cache intelligent** : Éviter de recloner le repo à chaque build
   - Solution : `git pull` au lieu de clone si le repo existe déjà
5. **Webhooks** : Notifier quand un build est terminé

   - Solution : Système de webhooks configurables par projet

6. **Build artifacts** : Stocker plusieurs fichiers (binaires + assets)
   - Solution : Archive tar.gz ou zip avec tous les fichiers nécessaires
