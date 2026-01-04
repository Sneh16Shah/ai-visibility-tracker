package services

import (
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/db"
	"github.com/Sneh16Shah/ai-visibility-tracker/models"
)

// MetricsCalculator handles all metrics calculations
type MetricsCalculator struct{}

// NewMetricsCalculator creates a new metrics calculator
func NewMetricsCalculator() *MetricsCalculator {
	return &MetricsCalculator{}
}

// CalculateAndStoreMetrics calculates all metrics for a brand and stores a snapshot
func (m *MetricsCalculator) CalculateAndStoreMetrics(brandID int) (*models.MetricSnapshot, error) {
	// Get all AI responses for this brand
	responseRepo := db.NewAIResponseRepository()
	responses, err := responseRepo.GetByBrandID(brandID)
	if err != nil {
		return nil, err
	}

	// Aggregate mention data
	mentionRepo := db.NewMentionRepository()

	var totalMentions int
	var brandMentions int
	var competitorMentions int
	var positiveCount int
	var neutralCount int
	var negativeCount int

	for _, response := range responses {
		mentions, err := mentionRepo.GetByResponseID(response.ID)
		if err != nil {
			continue
		}

		for _, mention := range mentions {
			totalMentions++

			if mention.EntityType == "brand" {
				brandMentions++
			} else {
				competitorMentions++
			}

			switch mention.Sentiment {
			case "positive":
				positiveCount++
			case "negative":
				negativeCount++
			default:
				neutralCount++
			}
		}
	}

	// Calculate metrics
	visibilityScore := m.calculateVisibilityScore(brandMentions, totalMentions, positiveCount, negativeCount)
	citationShare := m.calculateCitationShare(brandMentions, totalMentions)

	// Create snapshot
	snapshot := &models.MetricSnapshot{
		BrandID:         brandID,
		VisibilityScore: visibilityScore,
		CitationShare:   citationShare,
		MentionCount:    brandMentions,
		PositiveCount:   positiveCount,
		NeutralCount:    neutralCount,
		NegativeCount:   negativeCount,
		SnapshotDate:    time.Now(),
	}

	// Store snapshot
	metricRepo := db.NewMetricRepository()
	storedSnapshot, err := metricRepo.Create(snapshot)
	if err != nil {
		return nil, err
	}

	return storedSnapshot, nil
}

// calculateVisibilityScore calculates visibility score (0-100)
func (m *MetricsCalculator) calculateVisibilityScore(brandMentions, totalMentions, positive, negative int) float64 {
	if totalMentions == 0 {
		return 0
	}

	// Base score from citation share (0-50)
	citationScore := float64(brandMentions) / float64(totalMentions) * 50

	// Sentiment bonus/penalty (0-50)
	totalSentiment := positive + negative
	if totalSentiment == 0 {
		return citationScore + 25 // Neutral gets 25
	}

	sentimentRatio := float64(positive-negative) / float64(totalSentiment)
	sentimentScore := 25 + (sentimentRatio * 25) // Range: 0-50

	return citationScore + sentimentScore
}

// calculateCitationShare calculates citation share percentage
func (m *MetricsCalculator) calculateCitationShare(brandMentions, totalMentions int) float64 {
	if totalMentions == 0 {
		return 0
	}
	return float64(brandMentions) / float64(totalMentions) * 100
}

// GetDashboardMetrics returns aggregated metrics for the dashboard
func (m *MetricsCalculator) GetDashboardMetrics(brandID int) (*models.DashboardData, error) {
	metricRepo := db.NewMetricRepository()
	brandRepo := db.NewBrandRepository()

	// Get latest snapshot
	latest, err := metricRepo.GetLatestByBrandID(brandID)
	if err != nil {
		// Return empty data if no metrics
		return m.getEmptyDashboardData(), nil
	}

	// Get trends (last 7 days)
	trends, _ := metricRepo.GetTrendsByBrandID(brandID, 7)

	// Get brand info for competitor breakdown
	brand, err := brandRepo.GetByID(brandID)
	if err != nil {
		return nil, err
	}

	// Calculate citation breakdown
	citationBreakdown := m.calculateCitationBreakdown(brandID, brand)

	// Calculate competitor comparison
	competitorData := m.calculateCompetitorMetrics(brandID, brand)

	// Calculate sentiment score (1-5 scale)
	sentimentScore := m.calculateSentimentScore(latest.PositiveCount, latest.NeutralCount, latest.NegativeCount)

	return &models.DashboardData{
		VisibilityScore:   latest.VisibilityScore,
		CitationShare:     latest.CitationShare,
		TotalMentions:     latest.MentionCount,
		SentimentScore:    sentimentScore,
		Trends:            trends,
		CitationBreakdown: citationBreakdown,
		CompetitorData:    competitorData,
	}, nil
}

// calculateSentimentScore converts counts to 1-5 scale
func (m *MetricsCalculator) calculateSentimentScore(positive, neutral, negative int) float64 {
	total := positive + neutral + negative
	if total == 0 {
		return 3.0 // Neutral default
	}

	// Weighted: positive=5, neutral=3, negative=1
	score := float64(positive*5+neutral*3+negative*1) / float64(total)
	return score
}

// calculateCitationBreakdown calculates share for each entity
func (m *MetricsCalculator) calculateCitationBreakdown(brandID int, brand *models.Brand) []models.CitationBreakdown {
	// TODO: Get actual mention counts from database
	// For now, return placeholder data
	breakdown := []models.CitationBreakdown{
		{Name: brand.Name, Value: 35, Color: "#6366f1"},
	}

	colors := []string{"#10b981", "#f59e0b", "#ef4444", "#8b5cf6"}
	for i, comp := range brand.Competitors {
		if i >= len(colors) {
			break
		}
		breakdown = append(breakdown, models.CitationBreakdown{
			Name:  comp.Name,
			Value: float64(25 - i*5), // Descending values
			Color: colors[i],
		})
	}

	return breakdown
}

// calculateCompetitorMetrics calculates metrics for each competitor
func (m *MetricsCalculator) calculateCompetitorMetrics(brandID int, brand *models.Brand) []models.CompetitorMetrics {
	// TODO: Get actual competitor metrics from database
	// For now, return placeholder data
	metrics := []models.CompetitorMetrics{
		{Name: brand.Name, Mentions: 35, Positive: 28, Neutral: 5, Negative: 2},
	}

	for i, comp := range brand.Competitors {
		mentions := 28 - (i * 5)
		if mentions < 10 {
			mentions = 10
		}
		metrics = append(metrics, models.CompetitorMetrics{
			Name:     comp.Name,
			Mentions: mentions,
			Positive: int(float64(mentions) * 0.7),
			Neutral:  int(float64(mentions) * 0.2),
			Negative: int(float64(mentions) * 0.1),
		})
	}

	return metrics
}

// getEmptyDashboardData returns empty dashboard data
func (m *MetricsCalculator) getEmptyDashboardData() *models.DashboardData {
	return &models.DashboardData{
		VisibilityScore:   0,
		CitationShare:     0,
		TotalMentions:     0,
		SentimentScore:    3.0,
		Trends:            []models.MetricSnapshot{},
		CitationBreakdown: []models.CitationBreakdown{},
		CompetitorData:    []models.CompetitorMetrics{},
	}
}
