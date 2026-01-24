package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/db"
	"github.com/Sneh16Shah/ai-visibility-tracker/models"
	"github.com/Sneh16Shah/ai-visibility-tracker/services"
	"github.com/gin-gonic/gin"
)

// HealthCheck returns the health status of the API
func HealthCheck(c *gin.Context) {
	dbStatus := "disconnected"
	if db.GetDB() != nil {
		if err := db.GetDB().Ping(); err == nil {
			dbStatus = "connected"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "ai-visibility-tracker",
		"database":  dbStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ============================================
// Helper Functions
// ============================================

// getUserID extracts userID from context, defaults to 1 if not set
func getUserID(c *gin.Context) int {
	userID, exists := c.Get("userID")
	if !exists {
		return 1 // Default user for demo mode
	}
	return userID.(int)
}

// ============================================
// Brand Controllers
// ============================================

// GetBrands returns all brands for the current user
func GetBrands(c *gin.Context) {
	// Get userID from context (set by auth middleware)
	userID := getUserID(c)

	repo := db.NewBrandRepository()
	brands, err := repo.GetAll(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brands", "details": err.Error()})
		return
	}

	if brands == nil {
		brands = []models.Brand{}
	}

	c.JSON(http.StatusOK, gin.H{"brands": brands})
}

// CreateBrand creates a new brand
func CreateBrand(c *gin.Context) {
	var req models.CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Get userID from context (set by auth middleware)
	userID := getUserID(c)

	repo := db.NewBrandRepository()
	brand, err := repo.Create(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create brand", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, brand)
}

// GetBrand returns a single brand by ID
func GetBrand(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	repo := db.NewBrandRepository()
	brand, err := repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
		return
	}

	c.JSON(http.StatusOK, brand)
}

// UpdateBrand updates a brand
func UpdateBrand(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	var req models.UpdateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	repo := db.NewBrandRepository()
	brand, err := repo.Update(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, brand)
}

// DeleteBrand deletes a brand
func DeleteBrand(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	repo := db.NewBrandRepository()
	if err := repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete brand", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Brand deleted successfully"})
}

// ============================================
// Competitor Controllers
// ============================================

// GetCompetitors returns all competitors for a brand
func GetCompetitors(c *gin.Context) {
	brandID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	repo := db.NewBrandRepository()
	competitors, err := repo.GetCompetitors(brandID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch competitors", "details": err.Error()})
		return
	}

	if competitors == nil {
		competitors = []models.Competitor{}
	}

	c.JSON(http.StatusOK, gin.H{"competitors": competitors})
}

// AddCompetitor adds a competitor to a brand
func AddCompetitor(c *gin.Context) {
	brandID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	var req models.AddCompetitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	repo := db.NewBrandRepository()
	competitor, err := repo.AddCompetitor(brandID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add competitor", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, competitor)
}

// RemoveCompetitor removes a competitor
func RemoveCompetitor(c *gin.Context) {
	competitorID, err := strconv.Atoi(c.Param("competitorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid competitor ID"})
		return
	}

	repo := db.NewBrandRepository()
	if err := repo.RemoveCompetitor(competitorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove competitor", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Competitor removed successfully"})
}

// ============================================
// Alias Controllers
// ============================================

// GetAliases returns all aliases for a brand
func GetAliases(c *gin.Context) {
	brandID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	repo := db.NewBrandRepository()
	aliases, err := repo.GetAliases(brandID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch aliases", "details": err.Error()})
		return
	}

	if aliases == nil {
		aliases = []models.BrandAlias{}
	}

	c.JSON(http.StatusOK, gin.H{"aliases": aliases})
}

// AddAlias adds an alias to a brand
func AddAlias(c *gin.Context) {
	brandID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	var req models.AddAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	repo := db.NewBrandRepository()
	alias, err := repo.AddAlias(brandID, req.Alias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add alias", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, alias)
}

// RemoveAlias removes an alias
func RemoveAlias(c *gin.Context) {
	aliasID, err := strconv.Atoi(c.Param("aliasId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alias ID"})
		return
	}

	repo := db.NewBrandRepository()
	if err := repo.RemoveAlias(aliasID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove alias", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alias removed successfully"})
}

// UpdateAlertSettings updates alert threshold and schedule frequency for a brand
func UpdateAlertSettings(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	var req struct {
		AlertThreshold    float64 `json:"alert_threshold"`
		ScheduleFrequency string  `json:"schedule_frequency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Validate schedule frequency
	validFrequencies := map[string]bool{"disabled": true, "daily": true, "weekly": true}
	if !validFrequencies[req.ScheduleFrequency] {
		req.ScheduleFrequency = "disabled"
	}

	repo := db.NewBrandRepository()
	if err := repo.UpdateAlertSettings(id, req.AlertThreshold, req.ScheduleFrequency); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert settings updated successfully"})
}

// ============================================
// Prompt Controllers
// ============================================

// GetPrompts returns all active prompts
func GetPrompts(c *gin.Context) {
	repo := db.NewPromptRepository()
	prompts, err := repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prompts", "details": err.Error()})
		return
	}

	if prompts == nil {
		prompts = []models.Prompt{}
	}

	c.JSON(http.StatusOK, gin.H{"prompts": prompts})
}

// CreatePrompt creates a new prompt
func CreatePrompt(c *gin.Context) {
	var req struct {
		Category    string `json:"category" binding:"required"`
		Template    string `json:"template" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	repo := db.NewPromptRepository()
	prompt, err := repo.Create(req.Category, req.Template, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create prompt", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, prompt)
}

// DeletePrompt deletes a prompt by ID
func DeletePrompt(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompt ID"})
		return
	}

	repo := db.NewPromptRepository()
	if err := repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete prompt", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prompt deleted"})
}

// UpdatePrompt updates an existing prompt
func UpdatePrompt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompt ID"})
		return
	}

	var req struct {
		Category    string `json:"category"`
		Template    string `json:"template"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	repo := db.NewPromptRepository()
	prompt, err := repo.Update(id, req.Category, req.Template, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prompt", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prompt)
}

// ============================================
// Analysis Controllers
// ============================================

// GetAnalysisStatus returns the current status of the analysis service
func GetAnalysisStatus(c *gin.Context) {
	svc := services.GetAnalysisService()
	if svc == nil {
		c.JSON(http.StatusOK, gin.H{
			"provider_available": false,
			"can_run_analysis":   false,
			"message":            "Analysis service not initialized",
		})
		return
	}

	status := svc.GetStatus()
	c.JSON(http.StatusOK, status)
}

// RunAnalysis executes the analysis for a brand with rate limiting protection
func RunAnalysis(c *gin.Context) {
	var req models.RunAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	svc := services.GetAnalysisService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Analysis service not available",
			"message": "Please configure OPENAI_API_KEY or AI_PROVIDER=ollama",
		})
		return
	}

	// Check if we can run analysis (rate limit and in-flight check)
	canRun, reason := svc.CanRun(req.BrandID)
	if !canRun {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":           "Cannot run analysis",
			"reason":          reason,
			"retry_after_sec": 60, // Suggest retry after 1 minute
		})
		return
	}

	// Run the analysis
	ctx := c.Request.Context()
	result, err := svc.RunAnalysis(ctx, req.BrandID, req.PromptIDs)
	if err != nil {
		// Check for specific errors
		if err.Error() == "analysis already in progress for this brand" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Analysis already in progress",
				"message": "Please wait for the current analysis to complete",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Analysis failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAnalysisResults returns all analysis results for a brand
func GetAnalysisResults(c *gin.Context) {
	brandID, _ := strconv.Atoi(c.Query("brand_id"))
	if brandID == 0 {
		brandID = 1 // Default for demo
	}

	repo := db.NewAIResponseRepository()
	results, err := repo.GetByBrandID(brandID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results", "details": err.Error()})
		return
	}

	if results == nil {
		results = []models.AIResponse{}
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// GetAnalysisResult returns a single analysis result
func GetAnalysisResult(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid result ID"})
		return
	}

	repo := db.NewAIResponseRepository()
	result, err := repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ============================================
// Metrics Controllers
// ============================================

// GetMetrics returns metrics for a brand
func GetMetrics(c *gin.Context) {
	brandID, _ := strconv.Atoi(c.Query("brand_id"))
	if brandID == 0 {
		brandID = 1 // Default for demo
	}

	repo := db.NewMetricRepository()
	metrics, err := repo.GetTrendsByBrandID(brandID, 30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics", "details": err.Error()})
		return
	}

	if metrics == nil {
		metrics = []models.MetricSnapshot{}
	}

	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

// GetDashboardData returns aggregated dashboard data
func GetDashboardData(c *gin.Context) {
	brandID, _ := strconv.Atoi(c.Query("brand_id"))
	if brandID == 0 {
		brandID = 1 // Default for demo
	}

	metricRepo := db.NewMetricRepository()

	// Get latest metrics
	latest, err := metricRepo.GetLatestByBrandID(brandID)
	if err != nil {
		// Return demo data if no metrics exist
		c.JSON(http.StatusOK, getDemoData())
		return
	}

	// Get trends
	trends, _ := metricRepo.GetTrendsByBrandID(brandID, 7)

	c.JSON(http.StatusOK, models.DashboardData{
		VisibilityScore: latest.VisibilityScore,
		CitationShare:   latest.CitationShare,
		TotalMentions:   latest.MentionCount,
		SentimentScore:  calculateSentimentScore(latest),
		Trends:          trends,
	})
}

// Helper function to calculate sentiment score
func calculateSentimentScore(m *models.MetricSnapshot) float64 {
	total := m.PositiveCount + m.NeutralCount + m.NegativeCount
	if total == 0 {
		return 0
	}
	// Weighted score: positive=5, neutral=3, negative=1
	score := float64(m.PositiveCount*5+m.NeutralCount*3+m.NegativeCount*1) / float64(total)
	return score
}

// getDemoData returns empty data when no real data exists
func getDemoData() models.DashboardData {
	return models.DashboardData{
		VisibilityScore:   0,
		CitationShare:     0,
		TotalMentions:     0,
		SentimentScore:    0,
		Trends:            []models.MetricSnapshot{},
		CitationBreakdown: []models.CitationBreakdown{},
		CompetitorData:    []models.CompetitorMetrics{},
	}
}

// ============================================
// Export Controllers
// ============================================

// ExportCSV exports metrics data as CSV
func ExportCSV(c *gin.Context) {
	brandIDStr := c.Query("brand_id")
	if brandIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "brand_id is required"})
		return
	}

	brandID, err := strconv.Atoi(brandIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand_id"})
		return
	}

	// Get brand info
	brandRepo := db.NewBrandRepository()
	brand, err := brandRepo.GetByID(brandID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
		return
	}

	// Get metrics history (up to 365 days)
	metricsRepo := db.NewMetricRepository()
	snapshots, err := metricsRepo.GetTrendsByBrandID(brandID, 365)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}

	// Build CSV
	var csvContent strings.Builder
	csvContent.WriteString("Date,Visibility Score,Citation Share,Total Mentions,Positive,Neutral,Negative\n")

	for _, s := range snapshots {
		line := fmt.Sprintf("%s,%.1f,%.1f,%d,%d,%d,%d\n",
			s.CreatedAt.Format("2006-01-02 15:04"),
			s.VisibilityScore,
			s.CitationShare,
			s.MentionCount,
			s.PositiveCount,
			s.NeutralCount,
			s.NegativeCount,
		)
		csvContent.WriteString(line)
	}

	// Set headers for CSV download
	filename := fmt.Sprintf("%s_visibility_report_%s.csv", brand.Name, time.Now().Format("2006-01-02"))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.String(http.StatusOK, csvContent.String())
}
