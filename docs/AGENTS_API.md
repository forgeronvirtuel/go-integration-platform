# API Agents - Documentation

## Vue d'ensemble

L'API Agents permet de gérer les agents de build qui exécutent les tâches de compilation. Chaque agent a un nom unique (généralement le hostname), des labels pour le routage des builds, et un statut.

## Structure d'un Agent

```json
{
  "id": 1,
  "name": "build-agent-01",
  "labels": {
    "os": "linux",
    "arch": "amd64",
    "region": "eu-west",
    "gpu": "nvidia"
  },
  "status": "ONLINE",
  "last_seen_at": "2025-12-13T17:43:05Z",
  "created_at": "2025-12-13T16:42:39Z"
}
```

### Champs

- **id** (int) : Identifiant unique de l'agent
- **name** (string) : Nom de l'agent (généralement hostname), doit être unique
- **labels** (map[string]string) : Labels key/value pour le routage et la sélection
- **status** (string) : Statut de l'agent
  - `ONLINE` : Agent actif et disponible
  - `OFFLINE` : Agent hors ligne
  - `DRAINING` : Agent en cours de drainage (n'accepte plus de nouveaux builds)
- **last_seen_at** (datetime) : Dernier heartbeat reçu
- **created_at** (datetime) : Date de création

## Endpoints

### 1. Créer un agent

**POST** `/v1/api/agents`

Crée un nouvel agent. Le statut initial est `OFFLINE`.

**Request Body:**

```json
{
  "name": "build-agent-01",
  "labels": {
    "os": "linux",
    "arch": "amd64",
    "region": "eu-west"
  }
}
```

**Response:** `201 Created`

```json
{
  "id": 1,
  "name": "build-agent-01",
  "labels": {
    "os": "linux",
    "arch": "amd64",
    "region": "eu-west"
  },
  "status": "OFFLINE",
  "last_seen_at": "2025-12-13T17:00:00Z",
  "created_at": "2025-12-13T17:00:00Z"
}
```

**Erreurs:**

- `400 Bad Request` : Données invalides
- `409 Conflict` : Un agent avec ce nom existe déjà

**Exemple:**

```bash
curl -X POST http://localhost:3000/v1/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "build-agent-01",
    "labels": {
      "os": "linux",
      "arch": "amd64"
    }
  }'
```

---

### 2. Lister tous les agents

**GET** `/v1/api/agents`

Récupère la liste de tous les agents. Supporte un filtre optionnel par statut.

**Query Parameters:**

- `status` (optionnel) : Filtrer par statut (`ONLINE`, `OFFLINE`, `DRAINING`)

**Response:** `200 OK`

```json
{
  "agents": [
    {
      "id": 1,
      "name": "build-agent-01",
      "labels": { "os": "linux" },
      "status": "ONLINE",
      "last_seen_at": "2025-12-13T17:43:00Z",
      "created_at": "2025-12-13T17:00:00Z"
    },
    {
      "id": 2,
      "name": "build-agent-02",
      "labels": { "os": "darwin" },
      "status": "OFFLINE",
      "last_seen_at": "2025-12-13T16:30:00Z",
      "created_at": "2025-12-13T16:00:00Z"
    }
  ],
  "count": 2
}
```

**Exemples:**

```bash
# Tous les agents
curl http://localhost:3000/v1/api/agents

# Uniquement les agents ONLINE
curl http://localhost:3000/v1/api/agents?status=ONLINE

# Uniquement les agents OFFLINE
curl http://localhost:3000/v1/api/agents?status=OFFLINE
```

---

### 3. Récupérer un agent par ID

**GET** `/v1/api/agents/:id`

Récupère les détails d'un agent spécifique.

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "build-agent-01",
  "labels": { "os": "linux", "arch": "amd64" },
  "status": "ONLINE",
  "last_seen_at": "2025-12-13T17:43:00Z",
  "created_at": "2025-12-13T17:00:00Z"
}
```

**Erreurs:**

- `400 Bad Request` : ID invalide
- `404 Not Found` : Agent non trouvé

**Exemple:**

```bash
curl http://localhost:3000/v1/api/agents/1
```

---

### 4. Mettre à jour le statut d'un agent

**PUT** `/v1/api/agents/:id/status`

Change le statut d'un agent. Met également à jour `last_seen_at`.

**Request Body:**

```json
{
  "status": "ONLINE"
}
```

**Valeurs autorisées:** `ONLINE`, `OFFLINE`, `DRAINING`

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "build-agent-01",
  "labels": { "os": "linux" },
  "status": "ONLINE",
  "last_seen_at": "2025-12-13T17:45:00Z",
  "created_at": "2025-12-13T17:00:00Z"
}
```

**Erreurs:**

- `400 Bad Request` : ID ou statut invalide
- `404 Not Found` : Agent non trouvé

**Exemples:**

```bash
# Passer en ONLINE
curl -X PUT http://localhost:3000/v1/api/agents/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "ONLINE"}'

# Passer en DRAINING (maintenance)
curl -X PUT http://localhost:3000/v1/api/agents/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "DRAINING"}'

# Passer en OFFLINE
curl -X PUT http://localhost:3000/v1/api/agents/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "OFFLINE"}'
```

---

### 5. Mettre à jour les labels d'un agent

**PUT** `/v1/api/agents/:id/labels`

Met à jour les labels d'un agent. Remplace complètement les labels existants.

**Request Body:**

```json
{
  "labels": {
    "os": "linux",
    "arch": "arm64",
    "region": "us-east",
    "gpu": "nvidia-a100"
  }
}
```

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "build-agent-01",
  "labels": {
    "os": "linux",
    "arch": "arm64",
    "region": "us-east",
    "gpu": "nvidia-a100"
  },
  "status": "ONLINE",
  "last_seen_at": "2025-12-13T17:43:00Z",
  "created_at": "2025-12-13T17:00:00Z"
}
```

**Erreurs:**

- `400 Bad Request` : ID ou labels invalides
- `404 Not Found` : Agent non trouvé

**Exemple:**

```bash
curl -X PUT http://localhost:3000/v1/api/agents/1/labels \
  -H "Content-Type: application/json" \
  -d '{
    "labels": {
      "os": "linux",
      "arch": "arm64",
      "region": "us-east"
    }
  }'
```

---

### 6. Heartbeat

**POST** `/v1/api/agents/:id/heartbeat`

Enregistre un heartbeat pour un agent. Met à jour `last_seen_at` et passe automatiquement l'agent de `OFFLINE` à `ONLINE` si nécessaire.

**Response:** `200 OK`

```json
{
  "message": "Heartbeat registered",
  "last_seen_at": "2025-12-13T17:45:30Z"
}
```

**Erreurs:**

- `400 Bad Request` : ID invalide
- `404 Not Found` : Agent non trouvé

**Exemple:**

```bash
curl -X POST http://localhost:3000/v1/api/agents/1/heartbeat
```

**Usage:** Les agents doivent envoyer un heartbeat régulièrement (recommandé: toutes les 30-60 secondes).

---

### 7. Supprimer un agent

**DELETE** `/v1/api/agents/:id`

Supprime un agent de la base de données.

**Response:** `200 OK`

```json
{
  "message": "Agent deleted successfully"
}
```

**Erreurs:**

- `400 Bad Request` : ID invalide
- `404 Not Found` : Agent non trouvé

**Exemple:**

```bash
curl -X DELETE http://localhost:3000/v1/api/agents/1
```

---

## Cycle de vie d'un agent

### 1. Enregistrement

```bash
# L'agent s'enregistre au démarrage
curl -X POST http://localhost:3000/v1/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "'$(hostname)'",
    "labels": {
      "os": "linux",
      "arch": "amd64"
    }
  }'
```

### 2. Passage en ligne

```bash
# L'agent indique qu'il est prêt
curl -X PUT http://localhost:3000/v1/api/agents/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "ONLINE"}'
```

### 3. Heartbeats réguliers

```bash
# Toutes les 30 secondes
while true; do
  curl -X POST http://localhost:3000/v1/api/agents/1/heartbeat
  sleep 30
done
```

### 4. Drainage (maintenance)

```bash
# Avant une maintenance
curl -X PUT http://localhost:3000/v1/api/agents/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "DRAINING"}'
```

### 5. Arrêt propre

```bash
# Avant de s'arrêter
curl -X PUT http://localhost:3000/v1/api/agents/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "OFFLINE"}'
```

---

## Labels et sélection d'agents

Les labels permettent de router les builds vers des agents spécifiques :

**Exemples de labels courants:**

```json
{
  "os": "linux", // Système d'exploitation
  "arch": "amd64", // Architecture CPU
  "region": "eu-west", // Région géographique
  "gpu": "nvidia", // Présence de GPU
  "docker": "true", // Support Docker
  "k8s": "true", // Cluster Kubernetes
  "env": "production" // Environnement
}
```

**Scénarios d'usage:**

1. **Builds spécifiques à l'architecture:**

   - Agents avec `arch: amd64` pour builds x86_64
   - Agents avec `arch: arm64` pour builds ARM

2. **Builds nécessitant du GPU:**

   - Sélectionner les agents avec `gpu: nvidia`

3. **Isolation par environnement:**
   - Production: `env: production`
   - Staging: `env: staging`

---

## Détection des agents inactifs

La base de données inclut une fonction pour marquer automatiquement les agents inactifs comme `OFFLINE` :

```go
// Marquer comme OFFLINE les agents sans heartbeat depuis 5 minutes
count, err := database.MarkStaleAgentsOffline(db, 5*time.Minute)
```

**Recommandation:** Exécuter cette fonction périodiquement (ex: toutes les minutes) dans une goroutine de supervision.

---

## Exemples d'intégration

### Agent simple en bash

```bash
#!/bin/bash

API_URL="http://localhost:3000/v1/api/agents"
AGENT_NAME=$(hostname)

# 1. Enregistrement
AGENT_ID=$(curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$AGENT_NAME\",
    \"labels\": {
      \"os\": \"$(uname -s | tr '[:upper:]' '[:lower:]')\",
      \"arch\": \"$(uname -m)\"
    }
  }" | jq -r '.id')

echo "Agent registered with ID: $AGENT_ID"

# 2. Passage en ONLINE
curl -s -X PUT "$API_URL/$AGENT_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "ONLINE"}' > /dev/null

# 3. Boucle de heartbeat
trap "curl -s -X PUT '$API_URL/$AGENT_ID/status' \
  -H 'Content-Type: application/json' \
  -d '{\"status\": \"OFFLINE\"}' > /dev/null; exit" INT TERM

while true; do
  curl -s -X POST "$API_URL/$AGENT_ID/heartbeat" > /dev/null
  sleep 30
done
```

### Agent en Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

type Agent struct {
    ID     int               `json:"id"`
    Name   string            `json:"name"`
    Labels map[string]string `json:"labels"`
    Status string            `json:"status"`
}

func main() {
    apiURL := "http://localhost:3000/v1/api/agents"
    hostname, _ := os.Hostname()

    // 1. Enregistrement
    agent := Agent{
        Name: hostname,
        Labels: map[string]string{
            "os":   "linux",
            "arch": "amd64",
        },
    }

    body, _ := json.Marshal(agent)
    resp, _ := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
    json.NewDecoder(resp.Body).Decode(&agent)
    fmt.Printf("Agent registered with ID: %d\n", agent.ID)

    // 2. Passage en ONLINE
    statusURL := fmt.Sprintf("%s/%d/status", apiURL, agent.ID)
    statusData := map[string]string{"status": "ONLINE"}
    statusBody, _ := json.Marshal(statusData)
    req, _ := http.NewRequest("PUT", statusURL, bytes.NewBuffer(statusBody))
    req.Header.Set("Content-Type", "application/json")
    http.DefaultClient.Do(req)

    // 3. Heartbeats
    heartbeatURL := fmt.Sprintf("%s/%d/heartbeat", apiURL, agent.ID)
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        http.Post(heartbeatURL, "application/json", nil)
    }
}
```

---

## Base de données

### Structure de la table

```sql
CREATE TABLE agents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    labels TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'OFFLINE'
        CHECK(status IN ('ONLINE', 'OFFLINE', 'DRAINING')),
    last_seen_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_name ON agents(name);
```

### Contraintes

- **name** : Doit être unique
- **status** : Uniquement `ONLINE`, `OFFLINE`, ou `DRAINING`
- **labels** : Stocké en JSON

---

## Bonnes pratiques

1. **Heartbeat régulier:** Envoyer un heartbeat toutes les 30-60 secondes
2. **Gestion des erreurs:** Réessayer en cas d'échec du heartbeat
3. **Arrêt propre:** Toujours passer en `OFFLINE` avant de s'arrêter
4. **Labels cohérents:** Utiliser des labels standardisés dans toute l'infrastructure
5. **Monitoring:** Surveiller le nombre d'agents `ONLINE` vs `OFFLINE`
6. **Drainage:** Utiliser `DRAINING` pour les maintenances planifiées

---

## Codes d'erreur

- `200 OK` : Succès
- `201 Created` : Agent créé avec succès
- `400 Bad Request` : Données invalides
- `404 Not Found` : Agent non trouvé
- `409 Conflict` : Agent avec ce nom existe déjà
- `500 Internal Server Error` : Erreur serveur
