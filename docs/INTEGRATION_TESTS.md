# Tests d'Intégration - Workflow Complet

Ce document décrit les tests d'intégration end-to-end qui valident l'ensemble du workflow de build de la plateforme GIP.

## Vue d'ensemble

Les tests d'intégration dans `full_workflow_test.go` simulent un utilisateur complet qui :

1. Crée un projet Go local avec un repo Git
2. L'enregistre dans la plateforme
3. Déclenche un build
4. Télécharge le binaire généré
5. L'exécute et vérifie sa sortie

## Tests Implémentés

### 1. `TestFullBuildWorkflow` ✅

**Scénario**: Workflow complet avec un projet Go simple

**Étapes**:

1. Crée un projet Go avec `cmd/main.go` qui affiche "Hello World from GIP Build!"
2. Initialise un repo Git local avec commit
3. Enregistre le projet via API `POST /api/projects`
4. Lance le build via API `POST /api/builds/`
5. Télécharge le binaire via `GET /api/builds/:id/download`
6. Exécute le binaire et vérifie que la sortie contient le message attendu

**Durée**: ~7-8 secondes

**Points de validation**:

- Projet créé avec succès (201 Created)
- Build en statut "success"
- Binaire téléchargeable (200 OK)
- Binaire de taille > 0 (typiquement ~2.2 MB)
- Headers HTTP corrects (Content-Type, Content-Disposition)
- Binaire exécutable produit la bonne sortie

### 2. `TestFullBuildWorkflow_WithSubdir` ✅

**Scénario**: Build avec sous-répertoire (monorepo)

**Étapes**:

1. Crée un projet avec structure: `services/api/cmd/main.go`
2. Configure le projet avec `subdir: "services/api"`
3. Exécute le même workflow de build
4. Vérifie que le binaire du sous-répertoire fonctionne

**Durée**: ~7-8 secondes

**Points de validation**:

- Support des sous-répertoires
- go.mod dans le sous-répertoire
- Build et exécution réussis

### 3. `TestBuildWorkflow_FailureCases` ✅

**Scénario**: Cas d'erreur

#### 3.1. `NoMainGo`

- Projet sans fichier `cmd/main.go`
- Vérifie que le build échoue avec erreur 400
- Message d'erreur: "cmd/main.go not found"

#### 3.2. `CompilationError`

- Code Go avec erreur de syntaxe
- Vérifie que le build échoue avec erreur 400
- Les logs contiennent l'erreur de compilation

**Durée**: ~7 secondes (pour les 2 sous-tests)

### 4. Tests Auxiliaires

#### `TestIntegrationBuildAndDownload` ✅

Test simplifié qui vérifie le workflow sans exécuter le binaire

#### `TestBuildEndpoint_ProjectNotFound` ✅

Vérifie l'erreur 404 pour un projet inexistant

#### `TestDownloadBinary_BuildNotFound` ✅

Vérifie l'erreur 404 pour un build inexistant

## Exécution des Tests

### Tous les tests d'intégration

```bash
make test-integration
# ou
CGO_ENABLED=1 go test -v ./tests/integration/ -timeout 15m
```

### Tests du workflow complet uniquement

```bash
CGO_ENABLED=1 go test -v ./tests/integration/ -run TestFullBuildWorkflow -timeout 10m
```

### Test spécifique

```bash
CGO_ENABLED=1 go test -v ./tests/integration/ -run TestFullBuildWorkflow_WithSubdir
```

### Mode court (skip les tests longs)

```bash
CGO_ENABLED=1 go test -v ./tests/integration/ -short
```

## Prérequis

### Outils requis

- Go 1.21+ installé
- Git installé et accessible dans le PATH
- SQLite support (CGO_ENABLED=1)

### Permissions

- Accès en lecture/écriture dans `/tmp` pour les projets temporaires
- Accès au workspace de test `./test-workspace`

## Structure des Tests

```
tests/integration/
├── integration_test.go       # Setup du serveur de test
├── full_workflow_test.go     # Tests de workflow complet (NOUVEAU)
└── build_test.go             # Tests auxiliaires de build
```

### Serveur de Test

Le serveur de test est démarré une seule fois dans `TestMain()`:

- Port: 9999
- Base de données: `./test_integration.db`
- Workspace: `./test-workspace`
- Nettoyé après chaque run

## Détails Techniques

### Création de Projets Go Temporaires

Les tests créent de vrais projets Go dans `/tmp`:

```go
projectDir := t.TempDir() + "/hello-world-project"
```

Structure créée:

```
hello-world-project/
├── .git/                 # Repo Git initialisé
├── go.mod                # Module Go valide
└── cmd/
    └── main.go           # Code Go fonctionnel
```

### Initialisation Git

Chaque projet test initialise Git:

```bash
git init
git config user.email "test@example.com"
git config user.name "Test User"
git add .
git commit -m "Initial commit"
```

### Gestion des Fichiers Binaires

**Important**: Les binaires téléchargés doivent être fermés avant exécution:

```go
outFile, _ := os.Create(binaryPath)
io.Copy(outFile, resp.Body)
outFile.Close()  // ⚠️ Important pour éviter "text file busy"
os.Chmod(binaryPath, 0755)
time.Sleep(200 * time.Millisecond)  // Délai de sécurité
exec.Command(binaryPath).Run()
```

### Timeouts

- Timeout global des tests: 15 minutes
- Build individuel: 5 minutes (défini dans builds.go)
- Délai de stabilisation fichier: 200ms

## Statistiques

### Temps d'Exécution Typiques

- TestFullBuildWorkflow: ~7.5s
- TestFullBuildWorkflow_WithSubdir: ~7.5s
- TestBuildWorkflow_FailureCases: ~7s
- **Total**: ~23-25 secondes pour tous les tests

### Tailles de Binaires

- Binaire Hello World simple: ~2.2 MB
- Inclut le runtime Go complet

## Dépannage

### "text file busy"

**Symptôme**: Erreur lors de l'exécution du binaire téléchargé  
**Cause**: Fichier encore ouvert pour écriture  
**Solution**: Appeler `outFile.Close()` avant `exec.Command()`

### "Build timeout"

**Symptôme**: Build qui dépasse 5 minutes  
**Cause**: go mod download trop lent ou repo trop volumineux  
**Solution**: Augmenter le timeout dans builds.go ou optimiser les dépendances

### "cmd/main.go not found"

**Symptôme**: Build échoue même avec le bon fichier  
**Cause**: Structure du répertoire incorrecte ou subdir mal configuré  
**Solution**: Vérifier le chemin exact avec les logs de build

### Tests instables

**Symptôme**: Tests qui passent/échouent aléatoirement  
**Cause**: Race conditions ou fichiers non fermés  
**Solution**: Augmenter les `time.Sleep()` ou ajouter des synchronisations

## Améliorations Futures

### Court terme

- [ ] Parallélisation des tests (actuellement séquentiels)
- [ ] Cache des modules Go entre les tests
- [ ] Nettoyage automatique des builds anciens

### Moyen terme

- [ ] Tests de charge (builds simultanés)
- [ ] Tests de gros projets (>100MB)
- [ ] Validation des logs de build
- [ ] Tests de timeout

### Long terme

- [ ] Tests multi-plateformes (Linux, macOS, Windows)
- [ ] Tests de projets avec dépendances C
- [ ] Tests de projets CGO
- [ ] Snapshot testing des binaires

## Contribution

Pour ajouter un nouveau test:

1. Créer une fonction `Test*` dans `full_workflow_test.go`
2. Utiliser `t.TempDir()` pour les fichiers temporaires
3. Toujours vérifier le status HTTP ET le contenu de la réponse
4. Nettoyer les ressources (fichiers, connexions)
5. Ajouter un marqueur ✅ dans les logs de succès

Exemple:

```go
func TestNewFeature(t *testing.T) {
    // Setup
    tempDir := t.TempDir()

    // Test logic
    // ...

    // Assertions
    assert.Equal(t, expected, actual)

    // Success marker
    t.Log("✅ New feature test passed!")
}
```

## Références

- [Documentation Build API](./BUILD_API.md)
- [Documentation Build Feature](./BUILD_FEATURE.md)
- [Makefile - cible test-integration](../Makefile)
