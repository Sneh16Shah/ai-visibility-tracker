package routes

import (
	"github.com/Sneh16Shah/ai-visibility-tracker/controllers"
	"github.com/gin-gonic/gin"
)

// Setup configures all API routes
func Setup(router *gin.Engine) {
	// Health check
	router.GET("/health", controllers.HealthCheck)

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/signup", controllers.Signup)
			auth.POST("/login", controllers.Login)
		}

		// Protected user route
		api.GET("/me", controllers.AuthMiddleware(), controllers.GetMe)

		// Brand routes (with optional auth - falls back to user 1)
		brands := api.Group("/brands")
		brands.Use(controllers.OptionalAuthMiddleware())
		{
			brands.GET("", controllers.GetBrands)
			brands.POST("", controllers.CreateBrand)
			brands.GET("/:id", controllers.GetBrand)
			brands.PUT("/:id", controllers.UpdateBrand)
			brands.DELETE("/:id", controllers.DeleteBrand)

			// Competitor routes (nested under brands)
			brands.GET("/:id/competitors", controllers.GetCompetitors)
			brands.POST("/:id/competitors", controllers.AddCompetitor)
			brands.DELETE("/:id/competitors/:competitorId", controllers.RemoveCompetitor)

			// Alias routes (nested under brands)
			brands.GET("/:id/aliases", controllers.GetAliases)
			brands.POST("/:id/aliases", controllers.AddAlias)
			brands.DELETE("/:id/aliases/:aliasId", controllers.RemoveAlias)
			brands.PUT("/:id/alerts", controllers.UpdateAlertSettings)

			// Insights routes (competitor deep dive)
			brands.GET("/:id/insights", controllers.GetInsights)
			brands.PUT("/:id/insights", controllers.SaveInsights)
		}

		// Prompt routes
		prompts := api.Group("/prompts")
		{
			prompts.GET("", controllers.GetPrompts)
			prompts.POST("", controllers.CreatePrompt)
			prompts.PUT("/:id", controllers.UpdatePrompt)
			prompts.DELETE("/:id", controllers.DeletePrompt)
		}

		// Analysis routes
		analysis := api.Group("/analysis")
		{
			analysis.GET("/status", controllers.GetAnalysisStatus)
			analysis.POST("/run", controllers.RunAnalysis)
			analysis.GET("/results", controllers.GetAnalysisResults)
			analysis.GET("/results/:id", controllers.GetAnalysisResult)
		}

		// Metrics routes
		metrics := api.Group("/metrics")
		{
			metrics.GET("", controllers.GetMetrics)
			metrics.GET("/dashboard", controllers.GetDashboardData)
		}

		// Export routes
		export := api.Group("/export")
		{
			export.GET("/csv", controllers.ExportCSV)
		}
	}
}
