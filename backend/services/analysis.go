package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/ai"
	"github.com/Sneh16Shah/ai-visibility-tracker/config"
	"github.com/Sneh16Shah/ai-visibility-tracker/db"
	"github.com/Sneh16Shah/ai-visibility-tracker/models"
)

// AnalysisService handles AI analysis with rate limiting
type AnalysisService struct {
	provider        ai.Provider
	rateLimiter     *ai.RateLimiter
	inFlightTracker *ai.InFlightTracker
	cfg             *config.Config
}

// Global singleton for the service
var analysisService *AnalysisService

// InitAnalysisService initializes the analysis service
func InitAnalysisService(cfg *config.Config) *AnalysisService {
	var provider ai.Provider

	// Choose provider based on config
	switch cfg.AIProvider {
	case "ollama":
		provider = ai.NewOllamaProvider("http://localhost:11434", "llama2")
		log.Println("🤖 Using Ollama (local LLM) as AI provider")
	case "openai":
		if cfg.OpenAIKey != "" {
			provider = ai.NewOpenAIProvider(cfg.OpenAIKey)
			log.Println("🤖 Using OpenAI as AI provider")
		}
	case "gemini":
		if cfg.GeminiKey != "" {
			provider = ai.NewGeminiProvider(cfg.GeminiKey)
			log.Println("🤖 Using Google Gemini as AI provider")
		}
	case "groq":
		if cfg.GroqKey != "" {
			provider = ai.NewGroqProvider(cfg.GroqKey)
			log.Println("🤖 Using Groq (fast inference) as AI provider")
		}
	case "openrouter":
		if cfg.OpenRouterKey != "" {
			provider = ai.NewOpenRouterProvider(cfg.OpenRouterKey)
			log.Println("🤖 Using OpenRouter as AI provider")
		}
	}

	// Fallback: try OpenRouter, then Groq, then Gemini, then OpenAI if no provider set
	if provider == nil {
		if cfg.OpenRouterKey != "" {
			provider = ai.NewOpenRouterProvider(cfg.OpenRouterKey)
			log.Println("🤖 Using OpenRouter as AI provider (auto-detected)")
		} else if cfg.GroqKey != "" {
			provider = ai.NewGroqProvider(cfg.GroqKey)
			log.Println("🤖 Using Groq as AI provider (auto-detected)")
		} else if cfg.GeminiKey != "" {
			provider = ai.NewGeminiProvider(cfg.GeminiKey)
			log.Println("🤖 Using Google Gemini as AI provider (auto-detected)")
		} else if cfg.OpenAIKey != "" {
			provider = ai.NewOpenAIProvider(cfg.OpenAIKey)
			log.Println("🤖 Using OpenAI as AI provider (auto-detected)")
		} else {
			log.Println("⚠️ No AI provider configured (set OPENROUTER_API_KEY, GROQ_API_KEY, GEMINI_API_KEY, or OPENAI_API_KEY)")
		}
	}

	// Rate limiter: 2 second minimum between calls, max 10 calls per minute
	rateLimiter := ai.NewRateLimiter(2*time.Second, 10)

	// In-flight tracker with 5 minute timeout
	inFlightTracker := ai.NewInFlightTracker(5 * time.Minute)

	analysisService = &AnalysisService{
		provider:        provider,
		rateLimiter:     rateLimiter,
		inFlightTracker: inFlightTracker,
		cfg:             cfg,
	}

	return analysisService
}

// GetAnalysisService returns the singleton service instance
func GetAnalysisService() *AnalysisService {
	return analysisService
}

// AnalysisStatus represents the current status of analysis capabilities
type AnalysisStatus struct {
	ProviderAvailable bool                   `json:"provider_available"`
	ProviderName      string                 `json:"provider_name"`
	RateLimitStatus   map[string]interface{} `json:"rate_limit_status"`
	CanRunAnalysis    bool                   `json:"can_run_analysis"`
}

// GetStatus returns the current status of the analysis service
func (s *AnalysisService) GetStatus() AnalysisStatus {
	providerAvailable := s.provider != nil && s.provider.IsAvailable()
	providerName := ""
	if s.provider != nil {
		providerName = s.provider.GetModelName()
	}

	rateLimitStatus := s.rateLimiter.GetStatus()
	canRun := providerAvailable && rateLimitStatus["can_proceed"].(bool)

	return AnalysisStatus{
		ProviderAvailable: providerAvailable,
		ProviderName:      providerName,
		RateLimitStatus:   rateLimitStatus,
		CanRunAnalysis:    canRun,
	}
}

// CanRun checks if we can run an analysis
func (s *AnalysisService) CanRun(brandID int) (bool, string) {
	// Check if provider is available
	if s.provider == nil || !s.provider.IsAvailable() {
		return false, "AI provider not configured or unavailable"
	}

	// Check if analysis is already in flight for this brand
	if s.inFlightTracker.IsInFlight(brandID) {
		return false, "Analysis already in progress for this brand"
	}

	// Check rate limiter
	if !s.rateLimiter.CanProceed() {
		waitTime := s.rateLimiter.TimeUntilNextAllowed()
		return false, fmt.Sprintf("Rate limited. Please wait %d seconds", int(waitTime.Seconds()))
	}

	return true, ""
}

// RunAnalysisResult represents the result of running an analysis
type RunAnalysisResult struct {
	Success      bool                `json:"success"`
	Message      string              `json:"message"`
	ResponsesRun int                 `json:"responses_run"`
	Responses    []models.AIResponse `json:"responses,omitempty"`
	Errors       []string            `json:"errors,omitempty"`
}

// RunAnalysis executes AI analysis for a brand
func (s *AnalysisService) RunAnalysis(ctx context.Context, brandID int, promptIDs []int) (*RunAnalysisResult, error) {
	// Try to acquire in-flight slot
	if !s.inFlightTracker.TryAcquire(brandID) {
		return nil, ai.ErrRequestInFlight
	}
	defer s.inFlightTracker.Release(brandID)

	// Check rate limiter
	if !s.rateLimiter.CanProceed() {
		waitTime := s.rateLimiter.TimeUntilNextAllowed()
		return nil, fmt.Errorf("%w: wait %d seconds", ai.ErrRateLimited, int(waitTime.Seconds()))
	}

	// Get brand info for prompt context
	brandRepo := db.NewBrandRepository()
	brand, err := brandRepo.GetByID(brandID)
	if err != nil {
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}

	// Get prompts
	promptRepo := db.NewPromptRepository()
	var prompts []models.Prompt
	if len(promptIDs) > 0 {
		// Get specific prompts
		for _, id := range promptIDs {
			prompt, err := promptRepo.GetByID(id)
			if err == nil {
				prompts = append(prompts, *prompt)
			}
		}
	} else {
		// Get all active prompts
		prompts, err = promptRepo.GetAll()
		if err != nil {
			return nil, fmt.Errorf("failed to get prompts: %w", err)
		}
	}

	// Limit number of prompts to avoid excessive API calls
	maxPrompts := 6
	if len(prompts) > maxPrompts {
		prompts = prompts[:maxPrompts]
	}

	result := &RunAnalysisResult{
		Success: true,
	}

	responseRepo := db.NewAIResponseRepository()

	// Delete existing responses for this brand before running new analysis
	// This ensures we only keep the latest run data
	if err := responseRepo.DeleteByBrandID(brandID); err != nil {
		log.Printf("Warning: failed to delete old responses for brand %d: %v", brandID, err)
		// Continue anyway - not critical
	}

	// Process each prompt
	for _, prompt := range prompts {
		// Check rate limit before each call
		if !s.rateLimiter.CanProceed() {
			result.Errors = append(result.Errors, "Rate limit reached, stopping analysis")
			break
		}

		// Build the actual prompt with brand context
		actualPrompt := s.buildPromptWithContext(prompt.Template, brand)

		// Record the call
		s.rateLimiter.RecordCall()

		// Query AI
		responseText, err := s.provider.Query(ctx, actualPrompt)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Prompt %d failed: %s", prompt.ID, err.Error()))
			continue
		}

		// Store the response
		aiResponse, err := responseRepo.Create(brandID, prompt.ID, actualPrompt, responseText, s.provider.GetModelName())
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to store response: %s", err.Error()))
			continue
		}

		// Detect mentions in the response
		mentionDetector := NewMentionDetector()
		detectedMentions := mentionDetector.DetectMentions(responseText, brand)

		// Store mentions
		if len(detectedMentions) > 0 {
			storedMentions, err := mentionDetector.StoreMentions(aiResponse.ID, detectedMentions)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to store mentions: %s", err.Error()))
			} else {
				aiResponse.Mentions = storedMentions
			}
		}

		result.Responses = append(result.Responses, *aiResponse)
		result.ResponsesRun++

		// Small delay between API calls to be respectful
		time.Sleep(500 * time.Millisecond)
	}

	// Calculate and store metrics after all prompts are processed
	if result.ResponsesRun > 0 {
		metricsCalc := NewMetricsCalculator()
		_, err := metricsCalc.CalculateAndStoreMetrics(brandID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to calculate metrics: %s", err.Error()))
		}
	}

	if len(result.Errors) > 0 && result.ResponsesRun == 0 {
		result.Success = false
		result.Message = "All prompts failed"
	} else if len(result.Errors) > 0 {
		result.Message = fmt.Sprintf("Completed with %d errors", len(result.Errors))
	} else {
		result.Message = fmt.Sprintf("Successfully processed %d prompts", result.ResponsesRun)
	}

	return result, nil
}

// buildPromptWithContext replaces template variables with brand context
func (s *AnalysisService) buildPromptWithContext(template string, brand *models.Brand) string {
	result := template

	// Replace {brand} with brand name
	result = strings.ReplaceAll(result, "{brand}", brand.Name)

	// Replace {category} with industry
	result = strings.ReplaceAll(result, "{category}", brand.Industry)

	// Replace {competitor} with first competitor if exists
	if len(brand.Competitors) > 0 {
		result = strings.ReplaceAll(result, "{competitor}", brand.Competitors[0].Name)
	} else {
		result = strings.ReplaceAll(result, "{competitor}", "similar products")
	}

	// Replace {use_case} with a generic use case based on industry
	result = strings.ReplaceAll(result, "{use_case}", "general business use")

	return result
}
