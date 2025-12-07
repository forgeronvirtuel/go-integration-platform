package server

import (
	"database/sql"
	"net/http"
	"strconv"

	"forgeronvirtuel/gip/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type CreateProjectRequest struct {
	Name    string `json:"name" binding:"required"`
	RepoURL string `json:"repo_url" binding:"required"`
	Branch  string `json:"branch"`
	Subdir  string `json:"subdir"`
}

type UpdateProjectRequest struct {
	Name    string `json:"name" binding:"required"`
	RepoURL string `json:"repo_url" binding:"required"`
	Branch  string `json:"branch" binding:"required"`
	Subdir  string `json:"subdir"`
}

func setupProjectRoutes(router *gin.Engine, db *sql.DB) {
	projects := router.Group("/api/projects")
	{
		// GET /api/projects - Liste tous les projets
		projects.GET("", func(c *gin.Context) {
			allProjects, err := database.GetAllProjects(db)
			if err != nil {
				log.Error().Err(err).Msg("Erreur lors de la récupération des projets")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "unable to fetch projects",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"projects": allProjects,
				"count":    len(allProjects),
			})
		})

		// GET /api/projects/:id - Récupère un projet par ID
		projects.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid project id",
				})
				return
			}

			project, err := database.GetProjectByID(db, id)
			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "project not found",
					})
					return
				}
				log.Error().Err(err).Int("id", id).Msg("Erreur lors de la récupération du projet")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "unable to fetch project",
				})
				return
			}

			c.JSON(http.StatusOK, project)
		})

		// GET /api/projects/by-name/:name - Récupère un projet par nom
		projects.GET("/by-name/:name", func(c *gin.Context) {
			name := c.Param("name")

			project, err := database.GetProjectByName(db, name)
			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "project not found",
					})
					return
				}
				log.Error().Err(err).Str("name", name).Msg("Erreur lors de la récupération du projet")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "unable to fetch project",
				})
				return
			}

			c.JSON(http.StatusOK, project)
		})

		// POST /api/projects - Crée un nouveau projet
		projects.POST("", func(c *gin.Context) {
			var req CreateProjectRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid request body",
					"details": err.Error(),
				})
				return
			}

			// Valeur par défaut pour la branche
			if req.Branch == "" {
				req.Branch = "main"
			}

			project, err := database.CreateProject(db, req.Name, req.RepoURL, req.Branch, req.Subdir)
			if err != nil {
				log.Error().Err(err).Str("name", req.Name).Msg("Erreur lors de la création du projet")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "unable to create project",
					"details": err.Error(),
				})
				return
			}

			log.Info().Int("id", project.ID).Str("name", project.Name).Msg("Projet créé avec succès")
			c.JSON(http.StatusCreated, project)
		})

		// PUT /api/projects/:id - Met à jour un projet
		projects.PUT("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid project id",
				})
				return
			}

			var req UpdateProjectRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid request body",
					"details": err.Error(),
				})
				return
			}

			// Vérifier que le projet existe
			_, err = database.GetProjectByID(db, id)
			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "project not found",
					})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "unable to fetch project",
				})
				return
			}

			project, err := database.UpdateProject(db, id, req.Name, req.RepoURL, req.Branch, req.Subdir)
			if err != nil {
				log.Error().Err(err).Int("id", id).Msg("Erreur lors de la mise à jour du projet")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "unable to update project",
					"details": err.Error(),
				})
				return
			}

			log.Info().Int("id", project.ID).Str("name", project.Name).Msg("Projet mis à jour avec succès")
			c.JSON(http.StatusOK, project)
		})

		// DELETE /api/projects/:id - Supprime un projet
		projects.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid project id",
				})
				return
			}

			// Vérifier que le projet existe
			project, err := database.GetProjectByID(db, id)
			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "project not found",
					})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "unable to fetch project",
				})
				return
			}

			err = database.DeleteProject(db, id)
			if err != nil {
				log.Error().Err(err).Int("id", id).Msg("Erreur lors de la suppression du projet")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "unable to delete project",
				})
				return
			}

			log.Info().Int("id", id).Str("name", project.Name).Msg("Projet supprimé avec succès")
			c.JSON(http.StatusOK, gin.H{
				"message": "project deleted successfully",
				"id":      id,
			})
		})
	}
}
