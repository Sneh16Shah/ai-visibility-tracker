package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenRouterProvider implements the Provider interface for OpenRouter
type OpenRouterProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// OpenRouterModels contains the free models available for comparison
var OpenRouterModels = []struct {
	ID       string
	Name     string
	Provider string
	Color    string
}{
	{ID: "google/gemma-3-27b-it:free", Name: "Gemma 3 27B", Provider: "Google", Color: "#4285f4"},
	{ID: "meta-llama/llama-3.3-70b-instruct:free", Name: "Llama 3.3 70B", Provider: "Meta", Color: "#0668e1"},
	{ID: "qwen/qwen3-coder:free", Name: "Qwen3 Coder", Provider: "Qwen", Color: "#6366f1"},
	{ID: "tngtech/deepseek-r1t2-chimera:free", Name: "DeepSeek Chimera", Provider: "TNG", Color: "#00d4aa"},
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(apiKey string) *OpenRouterProvider {
	return &OpenRouterProvider{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1/chat/completions",
		model:   "google/gemini-2.0-flash-001", // Fast and capable default model
		client: &http.Client{
			Timeout: 120 * time.Second, // Longer timeout for free models
		},
	}
}

// OpenRouterRequest represents the request to OpenRouter API (OpenAI compatible)
type OpenRouterRequest struct {
	Model    string              `json:"model"`
	Messages []OpenRouterMessage `json:"messages"`
}

// OpenRouterMessage represents a chat message
type OpenRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents the response from OpenRouter API
type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Code    interface{} `json:"code"` // Can be string or number
	} `json:"error,omitempty"`
}

// Query sends a prompt to OpenRouter and returns the response (uses default model)
func (p *OpenRouterProvider) Query(ctx context.Context, prompt string) (string, error) {
	return p.QueryWithModel(ctx, prompt, p.model)
}

// QueryWithModel sends a prompt to OpenRouter with a specific model
func (p *OpenRouterProvider) QueryWithModel(ctx context.Context, prompt string, model string) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("OpenRouter API key not configured")
	}

	// Build request
	reqBody := OpenRouterRequest{
		Model: model,
		Messages: []OpenRouterMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("HTTP-Referer", "https://ai-visibility-tracker.local") // Required by OpenRouter
	req.Header.Set("X-Title", "AI Visibility Tracker")                    // Optional but recommended

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var orResp OpenRouterResponse
	if err := json.Unmarshal(body, &orResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error
	if orResp.Error != nil {
		return "", fmt.Errorf("OpenRouter API error: %s", orResp.Error.Message)
	}

	// Extract text from response
	if len(orResp.Choices) == 0 || orResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no response from OpenRouter")
	}

	return orResp.Choices[0].Message.Content, nil
}

// IsAvailable checks if the OpenRouter provider is configured
func (p *OpenRouterProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetModelName returns the model name
func (p *OpenRouterProvider) GetModelName() string {
	return "openrouter-" + p.model
}

// GetAPIKey returns the API key (for multi-model comparison service)
func (p *OpenRouterProvider) GetAPIKey() string {
	return p.apiKey
}
