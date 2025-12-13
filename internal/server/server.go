package server

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

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
func SetupRouter(db *sql.DB, workspace string) *gin.Engine {
	if workspace == "" {
		workspace = "./workspace"
	}
	router := gin.Default()

	router.Static("/static", "./web/static")
	router.StaticFile("/", "./web/index.html")

	v1 := router.Group("/v1")

	// Endpoint: Health with DB check
	healthHandler := &HealthHandler{DB: db}
	v1.GET("/health", healthHandler.Health)

	setupProjectRoutes(v1, db)
	setupBuildRoutes(v1, db, workspace)
	setupAgentRoutes(v1, db)

	return router
}

func Start(port string, db *sql.DB, workspace string) {
	gin.SetMode(gin.ReleaseMode)

	router := SetupRouter(db, workspace)

	log.Info().Str("port", port).Msg("Serveur HTTP démarré")
	if err := router.Run(":" + port); err != nil {
		log.Fatal().Err(err).Msg("Impossible de démarrer le serveur")
	}
}
