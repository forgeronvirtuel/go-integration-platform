package server

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func HelloWorld(c *gin.Context) {
	log.Info().Msg("Route / appelée")
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World!",
		"status":  "ok",
	})
}

type HealthHandler struct {
	DB *sql.DB
}

func (h *HealthHandler) Health(c *gin.Context) {
	if err := h.DB.Ping(); err != nil {
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
}

// SetupRouter crée et configure le router Gin avec toutes les routes
func SetupRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()

	// Endpoint: Hello World
	router.GET("/", HelloWorld)

	// Endpoint: Health with DB check
	healthHandler := &HealthHandler{DB: db}
	router.GET("/health", healthHandler.Health)

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
