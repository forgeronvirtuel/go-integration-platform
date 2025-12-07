package server

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// SetupRouter crée et configure le router Gin avec toutes les routes
func SetupRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()

	// Route Hello World
	router.GET("/", func(c *gin.Context) {
		log.Info().Msg("Route / appelée")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
			"status":  "ok",
		})
	})

	// Route de santé avec vérification de la DB
	router.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			log.Error().Err(err).Msg("Erreur lors du ping de la base de données")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
				"error":  "database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	// Exemple de route avec requête SQL
	router.GET("/users", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, email FROM users")
		if err != nil {
			log.Error().Err(err).Msg("Erreur lors de la récupération des utilisateurs")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "unable to fetch users",
			})
			return
		}
		defer rows.Close()

		var users []map[string]interface{}
		for rows.Next() {
			var id int
			var name, email string
			if err := rows.Scan(&id, &name, &email); err != nil {
				log.Error().Err(err).Msg("Erreur lors du scan des résultats")
				continue
			}
			users = append(users, map[string]interface{}{
				"id":    id,
				"name":  name,
				"email": email,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"users": users,
		})
	})

	setupProjectRoutes(router, db)

	return router
}

func Start(port string, db *sql.DB) {
	// Configuration de Gin en mode release
	gin.SetMode(gin.ReleaseMode)

	router := SetupRouter(db)

	log.Info().Str("port", port).Msg("Serveur HTTP démarré")
	if err := router.Run(":" + port); err != nil {
		log.Fatal().Err(err).Msg("Impossible de démarrer le serveur")
	}
}
