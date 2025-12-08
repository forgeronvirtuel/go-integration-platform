package workspacemanager

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func ValidateWorkspaceDir(path string) error {
	// Convertir en chemin absolu
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Vérifier si le répertoire existe
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Tenter de créer le répertoire s'il n'existe pas
			log.Info().Str("workspace", absPath).Msg("Le répertoire de workspace n'existe pas, création en cours")
			if err := os.MkdirAll(absPath, 0755); err != nil {
				return err
			}
			log.Info().Str("workspace", absPath).Msg("Répertoire de workspace créé avec succès")
			return nil
		}
		return err
	}

	// Vérifier que c'est bien un répertoire
	if !info.IsDir() {
		return os.ErrInvalid
	}

	// Vérifier les permissions en lecture
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	file.Close()

	// Vérifier les permissions en écriture en créant un fichier temporaire
	testFile := filepath.Join(absPath, ".gip_test_write")
	f, err := os.Create(testFile)
	if err != nil {
		return err
	}
	f.Close()
	os.Remove(testFile)

	return nil
}
