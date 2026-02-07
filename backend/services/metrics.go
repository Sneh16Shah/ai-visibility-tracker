package services

import (
	"log"
	"math"
	"strings"
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/db"
	"github.com/Sneh16Shah/ai-visibility-tracker/models"
)

// Score weights as per specification
const (
	WeightMentionRate = 0.40 // 40%
	WeightPosition    = 0.25 // 25%
	WeightRecommend   = 0.20 // 20%
	WeightSentiment   = 0.15 // 15%
)

// Position weights for mentions within a response
const (
	PositionFirst  = 1.0
	PositionSecond = 0.7
	PositionLater  = 0.4
)

// MetricsCalculator handles all metrics calculations
type MetricsCalculator struct{}

// NewMetricsCalculator creates a new metrics calculator
func NewMetricsCalculator() *MetricsCalculator {
	return &MetricsCalculator{}
}

// CalculateAndStoreMetrics calculates all metrics for a brand and stores a snapshot
func (m *MetricsCalculator) CalculateAndStoreMetrics(brandID int) (*models.MetricSnapshot, error) {
	// Get only the latest run AI responses for this brand (not historical)
	responseRepo := db.NewAIResponseRepository()
	responses, err := responseRepo.GetLatestRunByBrandID(brandID)
	if err != nil {
		return nil, err
	}

	totalResponses := len(responses)
	if totalResponses == 0 {
		// Return empty snapshot if no responses
		return m.createEmptySnapshot(brandID)
	}

	// Aggregate mention data across all responses
	mentionRepo := db.NewMentionRepository()

	var totalMentions int
	var brandMentions int
	var positiveCount int
	var neutralCount int
	var negativeCount int
	var responsesWithBrand int
	var responsesWithRecommendation int
	var totalPositionScore float64
	var brandSentimentSum float64
	var categorySentimentSum float64
	var categoryMentionCount int

	for _, response := range responses {
		mentions, err := mentionRepo.GetByResponseID(response.ID)
		if err != nil {
			continue
		}

		hasBrand := false
		hasRecommendation := false
		brandMentionInResponse := 0

		for _, mention := range mentions {
			totalMentions++

			// Calculate sentiment score (1=negative, 3=neutral, 5=positive)
			sentimentValue := 3.0
			switch mention.Sentiment {
			case "positive":
				sentimentValue = 5.0
			case "negative":
				sentimentValue = 1.0
			}

			if mention.EntityType == "brand" {
				brandMentions++
				hasBrand = true
				brandMentionInResponse++

				// Count sentiment for brand mentions only
				switch mention.Sentiment {
				case "positive":
					positiveCount++
				case "negative":
					negativeCount++
				default:
					neutralCount++
				}

				// Calculate position weight based on PositionRank
				switch mention.PositionRank {
				case 1:
					totalPositionScore += PositionFirst // 1.0
				case 2:
					totalPositionScore += PositionSecond // 0.7
				default:
					totalPositionScore += PositionLater // 0.4
				}

				// Check for recommendation
				if mention.IsRecommendation {
					hasRecommendation = true
				}

				// Track brand sentiment
				brandSentimentSum += sentimentValue
			} else {
				// Competitor mention - contributes to category average
				categorySentimentSum += sentimentValue
				categoryMentionCount++
			}
		}

		if hasBrand {
			responsesWithBrand++
		}
		if hasRecommendation {
			responsesWithRecommendation++
		}
	}

	// 1. Normalized Mention Rate (0-1): responses with brand / total responses
	normalizedMentionRate := 0.0
	if totalResponses > 0 {
		normalizedMentionRate = float64(responsesWithBrand) / float64(totalResponses)
	}

	// 2. Weighted Position Score (0-1): normalize position scores
	// Max possible = 1.0 per response, normalize to 0-1
	weightedPositionScore := 0.0
	if totalResponses > 0 {
		weightedPositionScore = totalPositionScore / float64(totalResponses)
		// Clamp to 0-1 (can exceed 1 if multiple brand mentions)
		if weightedPositionScore > 1.0 {
			weightedPositionScore = 1.0
		}
	}

	// 3. Recommendation Rate (0-1): responses with explicit recommendation / total
	recommendationRate := 0.0
	if totalResponses > 0 {
		recommendationRate = float64(responsesWithRecommendation) / float64(totalResponses)
	}

	// 4. Relative Sentiment Index (0-1)
	// Brand sentiment vs category average, normalized to 0-1
	brandAvgSentiment := 3.0 // Default neutral
	if brandMentions > 0 {
		brandAvgSentiment = brandSentimentSum / float64(brandMentions)
	}

	categoryAvgSentiment := 3.0 // Default neutral
	if categoryMentionCount > 0 {
		categoryAvgSentiment = categorySentimentSum / float64(categoryMentionCount)
	}

	// Calculate relative sentiment: difference ranges from -4 to +4
	// Normalize to 0-1: (diff + 4) / 8
	relativeDiff := brandAvgSentiment - categoryAvgSentiment
	relativeSentimentIndex := (relativeDiff + 4.0) / 8.0
	if relativeSentimentIndex < 0 {
		relativeSentimentIndex = 0
	}
	if relativeSentimentIndex > 1 {
		relativeSentimentIndex = 1
	}

	// 5. Composite Visibility Score (0-100)
	visibilityScore := (WeightMentionRate*normalizedMentionRate +
		WeightPosition*weightedPositionScore +
		WeightRecommend*recommendationRate +
		WeightSentiment*relativeSentimentIndex) * 100

	// 6. Citation/Response Share (percentage of responses mentioning brand)
	citationShare := normalizedMentionRate * 100

	// 7. Confidence Score (based on historical variance)
	confidenceScore, confidenceLevel := m.calculateConfidenceScore(brandID)

	// Create snapshot with all component scores
	snapshot := &models.MetricSnapshot{
		BrandID:         brandID,
		VisibilityScore: visibilityScore,
		CitationShare:   citationShare,
		MentionCount:    brandMentions,
		PositiveCount:   positiveCount,
		NeutralCount:    neutralCount,
		NegativeCount:   negativeCount,
		SnapshotDate:    time.Now(),

		// Component scores (0-1)
		NormalizedMentionRate:  normalizedMentionRate,
		WeightedPositionScore:  weightedPositionScore,
		RecommendationRate:     recommendationRate,
		RelativeSentimentIndex: relativeSentimentIndex,

		// Confidence
		ConfidenceScore: confidenceScore,
		ConfidenceLevel: confidenceLevel,

		// Metadata
		ResponseCount:        totalResponses,
		CategoryAvgSentiment: categoryAvgSentiment,
	}

	log.Printf("📊 Composite Score for brand %d: %.1f (MentionRate=%.2f, Position=%.2f, Recommend=%.2f, Sentiment=%.2f)",
		brandID, visibilityScore, normalizedMentionRate, weightedPositionScore, recommendationRate, relativeSentimentIndex)

	// Store snapshot
	metricRepo := db.NewMetricRepository()
	storedSnapshot, err := metricRepo.Create(snapshot)
	if err != nil {
		return nil, err
	}

	return storedSnapshot, nil
}

// createEmptySnapshot creates an empty metric snapshot for brands with no data
func (m *MetricsCalculator) createEmptySnapshot(brandID int) (*models.MetricSnapshot, error) {
	snapshot := &models.MetricSnapshot{
		BrandID:         brandID,
		VisibilityScore: 0,
		CitationShare:   0,
		MentionCount:    0,
		SnapshotDate:    time.Now(),
		ConfidenceLevel: "low",
	}

	metricRepo := db.NewMetricRepository()
	return metricRepo.Create(snapshot)
}

// calculateConfidenceScore calculates confidence based on historical score variance
// Confidence = 1 - (stdDev / mean)
func (m *MetricsCalculator) calculateConfidenceScore(brandID int) (float64, string) {
	metricRepo := db.NewMetricRepository()
	trends, err := metricRepo.GetTrendsByBrandID(brandID, 7) // Last 7 snapshots
	if err != nil || len(trends) < 3 {
		return 0.5, "medium" // Not enough data
	}

	// Calculate mean
	var sum float64
	for _, t := range trends {
		sum += t.VisibilityScore
	}
	mean := sum / float64(len(trends))

	if mean == 0 {
		return 0, "low"
	}

	// Calculate standard deviation
	var varianceSum float64
	for _, t := range trends {
		diff := t.VisibilityScore - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(trends))
	stdDev := math.Sqrt(variance)

	// Confidence = 1 - (coefficient of variation)
	confidence := 1 - (stdDev / mean)
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	// Qualitative level
	level := "medium"
	if confidence >= 0.8 {
		level = "high"
	} else if confidence < 0.5 {
		level = "low"
	}

	return confidence, level
}

// calculateCitationShare calculates citation share percentage (legacy support)
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

	// Calculate per-model visibility
	modelVisibility := m.calculateModelVisibility(brandID)

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
		ModelVisibility:   modelVisibility,

		// Component scores
		NormalizedMentionRate:  latest.NormalizedMentionRate,
		WeightedPositionScore:  latest.WeightedPositionScore,
		RecommendationRate:     latest.RecommendationRate,
		RelativeSentimentIndex: latest.RelativeSentimentIndex,

		// Confidence
		ConfidenceScore: latest.ConfidenceScore,
		ConfidenceLevel: latest.ConfidenceLevel,

		// Metadata
		ResponseCount:        latest.ResponseCount,
		CategoryAvgSentiment: latest.CategoryAvgSentiment,
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

// Model color mapping for known AI models
var modelColors = map[string]string{
	// OpenRouter models
	"Gemma 3 27B":      "#4285f4",
	"Llama 3.3 70B":    "#0668e1",
	"Qwen3 Coder":      "#6366f1",
	"DeepSeek Chimera": "#00d4aa",
	// Groq
	"Groq Llama 3.3": "#f55036",
	// Legacy/other models
	"gpt-4":       "#10a37f",
	"gpt-4-turbo": "#10a37f",
	"claude-3":    "#d4a574",
	"gemini-pro":  "#4285f4",
}

// calculateModelVisibility calculates visibility scores per AI model
func (m *MetricsCalculator) calculateModelVisibility(brandID int) []models.ModelVisibility {
	responseRepo := db.NewAIResponseRepository()
	mentionRepo := db.NewMentionRepository()

	// Get all responses for this brand
	responses, err := responseRepo.GetByBrandID(brandID)
	if err != nil || len(responses) == 0 {
		return []models.ModelVisibility{}
	}

	// Group responses by model and calculate average scores
	type modelData struct {
		totalScore    int // Sum of all response scores
		responseCount int // Number of responses
		mentions      int // Total brand mentions
	}
	modelStats := make(map[string]*modelData)

	// Get brand info for score calculation
	brandRepo := db.NewBrandRepository()
	brand, err := brandRepo.GetByID(brandID)
	if err != nil {
		log.Printf("calculateModelVisibility: could not get brand: %v", err)
		return []models.ModelVisibility{}
	}

	for _, resp := range responses {
		modelName := resp.ModelName
		if modelName == "" {
			modelName = "Unknown"
		}

		if modelStats[modelName] == nil {
			modelStats[modelName] = &modelData{}
		}

		// Get mentions for this response
		mentions, err := mentionRepo.GetByResponseID(resp.ID)
		if err != nil {
			// Still count the response but with score 0
			modelStats[modelName].responseCount++
			continue
		}

		// Calculate score for THIS response using same logic as compare.go
		score := calculateResponseScore(mentions, brand.Name)
		modelStats[modelName].totalScore += score
		modelStats[modelName].responseCount++

		// Count brand mentions
		for _, mention := range mentions {
			if mention.EntityType == "brand" {
				modelStats[modelName].mentions++
			}
		}
	}

	// Convert to ModelVisibility slice with averaged scores
	var result []models.ModelVisibility
	for modelName, stats := range modelStats {
		if stats.responseCount == 0 {
			continue
		}

		// Calculate average score
		avgScore := float64(stats.totalScore) / float64(stats.responseCount)

		// Get color for this model
		color := "#888888" // Default gray
		for key, c := range modelColors {
			if strings.EqualFold(modelName, key) || strings.Contains(strings.ToLower(modelName), strings.ToLower(key)) {
				color = c
				break
			}
		}

		log.Printf("calculateModelVisibility: model=%s, responses=%d, totalScore=%d, avgScore=%.1f",
			modelName, stats.responseCount, stats.totalScore, avgScore)

		result = append(result, models.ModelVisibility{
			Model:    modelName,
			ModelID:  modelName,
			Color:    color,
			Score:    avgScore,
			Mentions: stats.mentions,
		})
	}

	log.Printf("calculateModelVisibility: returning %d model visibility entries", len(result))
	return result
}

// contains checks if s contains substr (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0)
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
		ModelVisibility:   []models.ModelVisibility{},
	}
}

// calculateResponseScore calculates the visibility score for a single response
// This uses the same logic as calculateVisibilityScore in compare.go
func calculateResponseScore(mentions []models.Mention, brandName string) int {
	var brandMention *models.Mention
	competitorCount := 0

	for i := range mentions {
		if mentions[i].EntityType == "brand" {
			brandMention = &mentions[i]
		} else if mentions[i].EntityType == "competitor" {
			competitorCount++
		}
	}

	if brandMention == nil {
		return 0
	}

	// Base score: 50 for being mentioned
	score := 50

	// Sentiment bonus
	switch brandMention.Sentiment {
	case "positive":
		score += 25
	case "negative":
		score -= 25
	}

	// Citation share bonus
	totalMentions := 1 + competitorCount
	citationShare := (1.0 / float64(totalMentions)) * 25
	score += int(citationShare)

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}
