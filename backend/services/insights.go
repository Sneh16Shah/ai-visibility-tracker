package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Sneh16Shah/ai-visibility-tracker/ai"
	"github.com/Sneh16Shah/ai-visibility-tracker/db"
)

// InsightsService handles AI-powered insights generation
type InsightsService struct {
	provider ai.Provider
}

// NewInsightsService creates a new insights service
func NewInsightsService() *InsightsService {
	// Create Gemini provider with API key from environment
	// Check both GEMINI_API_KEY (docker-compose) and GOOGLE_API_KEY (direct)
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	return &InsightsService{
		provider: ai.NewGeminiProvider(apiKey),
	}
}

// CompetitorInsightsResult represents the result of competitor analysis
type CompetitorInsightsResult struct {
	Success  bool   `json:"success"`
	Insights string `json:"insights"`
	Error    string `json:"error,omitempty"`
}

// GenerateCompetitorInsights generates AI-powered insights about competitors
func (s *InsightsService) GenerateCompetitorInsights(ctx context.Context, brandID int) (*CompetitorInsightsResult, error) {
	log.Printf("🔍 GenerateCompetitorInsights: Starting for brand %d", brandID)

	// Get brand info
	brandRepo := db.NewBrandRepository()
	brand, err := brandRepo.GetByID(brandID)
	if err != nil {
		return &CompetitorInsightsResult{
			Success: false,
			Error:   fmt.Sprintf("Brand not found: %v", err),
		}, nil
	}

	// Get competitors
	competitors := []string{}
	if brand.Competitors != nil {
		for _, c := range brand.Competitors {
			competitors = append(competitors, c.Name)
		}
	}

	if len(competitors) == 0 {
		return &CompetitorInsightsResult{
			Success: false,
			Error:   "No competitors configured for this brand",
		}, nil
	}

	// Check if AI provider is available
	if s.provider == nil || !s.provider.IsAvailable() {
		return &CompetitorInsightsResult{
			Success: false,
			Error:   "AI provider not configured. Please set GOOGLE_API_KEY.",
		}, nil
	}

	log.Printf("🔍 GenerateCompetitorInsights: Analyzing %s vs %v", brand.Name, competitors)

	// Build prompt
	prompt := fmt.Sprintf(`You are an AI visibility optimization expert. Analyze why competitors (%s) might rank better than "%s" in AI assistant responses.

Format your response with these sections:

## Why Competitors Rank Higher
- List 3-4 key reasons with specific examples

## Actionable Recommendations for %s
- List 5 specific, actionable steps to improve AI visibility
- Include SEO, content strategy, structured data, and brand authority tips

Keep each point concise (1-2 sentences). Industry: %s`,
		strings.Join(competitors, ", "),
		brand.Name,
		brand.Name,
		getIndustry(brand.Industry),
	)

	// Query AI
	response, err := s.provider.Query(ctx, prompt)
	if err != nil {
		log.Printf("🔍 GenerateCompetitorInsights: AI query failed: %v", err)
		return &CompetitorInsightsResult{
			Success: false,
			Error:   fmt.Sprintf("AI analysis failed: %v", err),
		}, nil
	}

	log.Printf("🔍 GenerateCompetitorInsights: Successfully generated insights for %s", brand.Name)

	return &CompetitorInsightsResult{
		Success:  true,
		Insights: response,
	}, nil
}

func getIndustry(industry string) string {
	if industry == "" {
		return "Technology"
	}
	return industry
}
