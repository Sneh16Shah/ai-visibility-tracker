package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/ai"
	"github.com/Sneh16Shah/ai-visibility-tracker/config"
	"github.com/Sneh16Shah/ai-visibility-tracker/db"
	"github.com/Sneh16Shah/ai-visibility-tracker/models"
)

// CompareService handles multi-model comparison via OpenRouter and Groq
type CompareService struct {
	openRouterProvider *ai.OpenRouterProvider
	groqProvider       *ai.GroqProvider
	cfg                *config.Config
}

// Global singleton for the compare service
var compareService *CompareService

// Groq model info (used when Groq API key is configured)
var GroqModelInfo = struct {
	ID       string
	Name     string
	Provider string
	Color    string
}{
	ID:       "groq",
	Name:     "Groq Llama 3.3",
	Provider: "Groq",
	Color:    "#f55036",
}

// InitCompareService initializes the compare service
func InitCompareService(cfg *config.Config) *CompareService {
	if cfg.OpenRouterKey == "" && cfg.GroqKey == "" {
		log.Println("⚠️ Neither OpenRouter nor Groq configured - Compare Models feature will be unavailable")
		return nil
	}

	compareService = &CompareService{
		cfg: cfg,
	}

	if cfg.OpenRouterKey != "" {
		compareService.openRouterProvider = ai.NewOpenRouterProvider(cfg.OpenRouterKey)
		log.Println("✅ Compare Models: OpenRouter enabled")
	}

	if cfg.GroqKey != "" {
		compareService.groqProvider = ai.NewGroqProvider(cfg.GroqKey)
		log.Println("✅ Compare Models: Groq enabled")
	}

	log.Println("✅ Compare Models service initialized")
	return compareService
}

// GetCompareService returns the singleton service instance
func GetCompareService() *CompareService {
	return compareService
}

// CompareModelsRequest represents the request for multi-model comparison
type CompareModelsRequest struct {
	BrandID   int      `json:"brand_id"`
	PromptIDs []int    `json:"prompt_ids"`
	ModelIDs  []string `json:"model_ids"` // Model IDs (OpenRouter or "groq")
}

// ModelResult represents a single model's response
type ModelResult struct {
	ModelID    string           `json:"model_id"`
	ModelName  string           `json:"model_name"`
	Provider   string           `json:"provider"`
	Color      string           `json:"color"`
	PromptText string           `json:"prompt_text"`
	Response   string           `json:"response"`
	Mentions   []models.Mention `json:"mentions"`
	Score      int              `json:"score"`
	Error      string           `json:"error,omitempty"`
	Timestamp  time.Time        `json:"timestamp"`
}

// CompareModelsResult represents the result of multi-model comparison
type CompareModelsResult struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	Results      []ModelResult `json:"results"`
	TotalCalls   int           `json:"total_calls"`
	SuccessCalls int           `json:"success_calls"`
	Errors       []string      `json:"errors,omitempty"`
}

// GetAvailableModels returns the list of available models for comparison
func (s *CompareService) GetAvailableModels() []map[string]string {
	var allModels []map[string]string

	// Add OpenRouter models if available
	if s.openRouterProvider != nil && s.openRouterProvider.IsAvailable() {
		for _, m := range ai.OpenRouterModels {
			allModels = append(allModels, map[string]string{
				"id":       m.ID,
				"name":     m.Name,
				"provider": m.Provider,
				"color":    m.Color,
			})
		}
	}

	// Add Groq if available
	if s.groqProvider != nil && s.groqProvider.IsAvailable() {
		allModels = append(allModels, map[string]string{
			"id":       GroqModelInfo.ID,
			"name":     GroqModelInfo.Name,
			"provider": GroqModelInfo.Provider,
			"color":    GroqModelInfo.Color,
		})
	}

	return allModels
}

// IsAvailable checks if the compare service is available
func (s *CompareService) IsAvailable() bool {
	if s == nil {
		return false
	}
	hasOpenRouter := s.openRouterProvider != nil && s.openRouterProvider.IsAvailable()
	hasGroq := s.groqProvider != nil && s.groqProvider.IsAvailable()
	return hasOpenRouter || hasGroq
}

// RunComparison runs multi-model comparison for the given prompts
func (s *CompareService) RunComparison(ctx context.Context, req CompareModelsRequest) (*CompareModelsResult, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("compare service not available - configure OPENROUTER_API_KEY or GROQ_API_KEY")
	}

	// Get brand info
	brandRepo := db.NewBrandRepository()
	brand, err := brandRepo.GetByID(req.BrandID)
	if err != nil {
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}

	// Get prompts
	promptRepo := db.NewPromptRepository()
	var prompts []models.Prompt
	if len(req.PromptIDs) > 0 {
		for _, id := range req.PromptIDs {
			prompt, err := promptRepo.GetByID(id)
			if err == nil {
				prompts = append(prompts, *prompt)
			}
		}
	} else {
		prompts, err = promptRepo.GetAll()
		if err != nil {
			return nil, fmt.Errorf("failed to get prompts: %w", err)
		}
	}

	// Limit prompts to avoid rate limiting
	maxPrompts := 6
	if len(prompts) > maxPrompts {
		prompts = prompts[:maxPrompts]
	}

	// Use default models if none specified (include Groq if available)
	modelIDs := req.ModelIDs
	if len(modelIDs) == 0 {
		if s.openRouterProvider != nil && s.openRouterProvider.IsAvailable() {
			for _, m := range ai.OpenRouterModels {
				modelIDs = append(modelIDs, m.ID)
			}
		}
		if s.groqProvider != nil && s.groqProvider.IsAvailable() {
			modelIDs = append(modelIDs, GroqModelInfo.ID)
		}
	}

	result := &CompareModelsResult{
		Success:    true,
		TotalCalls: len(prompts) * len(modelIDs),
	}

	// Create a mutex for thread-safe result appending
	var mu sync.Mutex
	var wg sync.WaitGroup

	mentionDetector := NewMentionDetector()

	// Process each prompt with each model (concurrently per model, sequentially per prompt)
	for _, prompt := range prompts {
		// Build actual prompt with brand context
		actualPrompt := buildPromptWithContext(prompt.Template, brand)

		// Query all models concurrently for this prompt
		for _, modelID := range modelIDs {
			wg.Add(1)
			go func(modelID string, prompt models.Prompt, actualPrompt string) {
				defer wg.Done()

				// Find model info - check if it's Groq first
				var modelName, provider, color string
				var response string
				var queryErr error

				if modelID == GroqModelInfo.ID {
					// Use Groq provider directly
					modelName = GroqModelInfo.Name
					provider = GroqModelInfo.Provider
					color = GroqModelInfo.Color

					if s.groqProvider != nil && s.groqProvider.IsAvailable() {
						response, queryErr = s.groqProvider.Query(ctx, actualPrompt)
					} else {
						queryErr = fmt.Errorf("Groq provider not available")
					}
				} else {
					// Use OpenRouter for other models
					for _, m := range ai.OpenRouterModels {
						if m.ID == modelID {
							modelName = m.Name
							provider = m.Provider
							color = m.Color
							break
						}
					}
					if modelName == "" {
						modelName = modelID
						provider = "Unknown"
						color = "#888888"
					}

					if s.openRouterProvider != nil && s.openRouterProvider.IsAvailable() {
						response, queryErr = s.openRouterProvider.QueryWithModel(ctx, actualPrompt, modelID)
					} else {
						queryErr = fmt.Errorf("OpenRouter provider not available")
					}
				}

				modelResult := ModelResult{
					ModelID:    modelID,
					ModelName:  modelName,
					Provider:   provider,
					Color:      color,
					PromptText: actualPrompt,
					Timestamp:  time.Now(),
				}

				if queryErr != nil {
					modelResult.Error = queryErr.Error()
					mu.Lock()
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", modelName, queryErr.Error()))
					result.Results = append(result.Results, modelResult)
					mu.Unlock()
					return
				}

				modelResult.Response = response

				// Detect mentions
				detectedMentions := mentionDetector.DetectMentions(response, brand)
				modelResult.Mentions = convertToModelMentions(detectedMentions)

				// Calculate score
				modelResult.Score = calculateVisibilityScore(modelResult.Mentions, brand.Name)

				mu.Lock()
				result.Results = append(result.Results, modelResult)
				result.SuccessCalls++
				mu.Unlock()

				// Small delay to avoid hitting rate limits
				time.Sleep(200 * time.Millisecond)
			}(modelID, prompt, actualPrompt)
		}

		// Wait for all models to respond for this prompt before moving to next
		wg.Wait()

		// Additional delay between prompts
		time.Sleep(500 * time.Millisecond)
	}

	if result.SuccessCalls == 0 && len(result.Errors) > 0 {
		result.Success = false
		result.Message = "All model queries failed"
	} else if len(result.Errors) > 0 {
		result.Message = fmt.Sprintf("Completed with %d/%d successful calls", result.SuccessCalls, result.TotalCalls)
	} else {
		result.Message = fmt.Sprintf("Successfully compared %d models across %d prompts", len(modelIDs), len(prompts))
	}

	// Store results to database for Dashboard display
	if result.SuccessCalls > 0 {
		s.storeCompareResults(req.BrandID, result)
	}

	return result, nil
}

// storeCompareResults saves compare results as AI responses for Dashboard visibility
func (s *CompareService) storeCompareResults(brandID int, result *CompareModelsResult) {
	log.Printf("📊 storeCompareResults: Starting for brand %d with %d results", brandID, len(result.Results))

	responseRepo := db.NewAIResponseRepository()
	mentionRepo := db.NewMentionRepository()
	promptRepo := db.NewPromptRepository()

	// Delete old responses for this brand before storing new ones
	if err := responseRepo.DeleteByBrandID(brandID); err != nil {
		log.Printf("Warning: failed to delete old responses for brand %d: %v", brandID, err)
	}

	// Get brand info for mention detection
	brandRepo := db.NewBrandRepository()
	brand, err := brandRepo.GetByID(brandID)
	if err != nil {
		log.Printf("Warning: failed to get brand info: %v", err)
		return
	}

	mentionDetector := NewMentionDetector()
	storedCount := 0
	mentionCount := 0

	// Store each successful result
	for _, modelResult := range result.Results {
		if modelResult.Error != "" {
			log.Printf("📊 Skipping model %s due to error: %s", modelResult.ModelName, modelResult.Error)
			continue // Skip failed results
		}

		// Find or use a default prompt ID
		prompts, err := promptRepo.GetAll()
		promptID := 1
		if err == nil && len(prompts) > 0 {
			promptID = prompts[0].ID
		}

		// Store the response with the model name
		storedResponse, err := responseRepo.Create(brandID, promptID, modelResult.PromptText, modelResult.Response, modelResult.ModelName)
		if err != nil {
			log.Printf("Warning: failed to store response for model %s: %v", modelResult.ModelName, err)
			continue
		}
		storedCount++
		log.Printf("📊 Stored response %d for model: %s", storedResponse.ID, modelResult.ModelName)

		// Detect and store mentions for this response
		detectedMentions := mentionDetector.DetectMentions(modelResult.Response, brand)
		log.Printf("📊 Detected %d mentions in response for model %s", len(detectedMentions), modelResult.ModelName)

		for _, mention := range detectedMentions {
			_, err := mentionRepo.Create(
				storedResponse.ID,
				mention.EntityName,
				string(mention.EntityType),
				string(mention.Sentiment),
				mention.ContextSnippet,
				mention.Position,
				mention.IsRecommendation,
				mention.PositionRank,
			)
			if err != nil {
				log.Printf("Warning: failed to store mention: %v", err)
			} else {
				mentionCount++
			}
		}
	}

	log.Printf("📊 storeCompareResults: Stored %d responses and %d mentions for brand %d", storedCount, mentionCount, brandID)

	// Recalculate metrics
	metricsCalc := NewMetricsCalculator()
	_, err = metricsCalc.CalculateAndStoreMetrics(brandID)
	if err != nil {
		log.Printf("Warning: failed to calculate metrics after compare: %v", err)
	}
	log.Printf("📊 storeCompareResults: Completed for brand %d", brandID)
}

// Helper to build prompt with brand context (reuse from analysis service)
func buildPromptWithContext(template string, brand *models.Brand) string {
	result := template

	// Replace {brand} with brand name
	result = replaceAllCaseInsensitive(result, "{brand}", brand.Name)

	// Replace {category} with industry
	result = replaceAllCaseInsensitive(result, "{category}", brand.Industry)

	// Replace {competitor} with first competitor if exists
	if len(brand.Competitors) > 0 {
		result = replaceAllCaseInsensitive(result, "{competitor}", brand.Competitors[0].Name)
	} else {
		result = replaceAllCaseInsensitive(result, "{competitor}", "similar products")
	}

	// Replace {use_case} with a generic use case
	result = replaceAllCaseInsensitive(result, "{use_case}", "general business use")

	return result
}

// Helper for case-insensitive replace
func replaceAllCaseInsensitive(s, old, new string) string {
	result := s
	for i := 0; i < len(result); i++ {
		if i+len(old) <= len(result) {
			if equalFold(result[i:i+len(old)], old) {
				result = result[:i] + new + result[i+len(old):]
				i += len(new) - 1
			}
		}
	}
	return result
}

func equalFold(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		c1, c2 := s1[i], s2[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}

// Convert detected mentions to model mentions format
func convertToModelMentions(detected []DetectedMention) []models.Mention {
	mentions := make([]models.Mention, len(detected))
	for i, d := range detected {
		mentions[i] = models.Mention{
			EntityName:     d.EntityName,
			EntityType:     d.EntityType,
			Sentiment:      string(d.Sentiment),
			ContextSnippet: d.ContextSnippet,
		}
	}
	return mentions
}

// Calculate visibility score based on mentions
func calculateVisibilityScore(mentions []models.Mention, brandName string) int {
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
