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

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// OpenAIRequest represents the request body for OpenAI API
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}

// OpenAIMessage represents a message in the OpenAI chat format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  "gpt-3.5-turbo", // Use cheaper model for this project
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.openai.com/v1/chat/completions",
	}
}

// IsAvailable checks if the provider is properly configured
func (p *OpenAIProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetModelName returns the model name
func (p *OpenAIProvider) GetModelName() string {
	return p.model
}

// Query sends a prompt to OpenAI and returns the response
func (p *OpenAIProvider) Query(ctx context.Context, prompt string) (string, error) {
	if !p.IsAvailable() {
		return "", ErrProviderNotReady
	}

	// Build request
	reqBody := OpenAIRequest{
		Model: p.model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant providing information about software tools and products. Give concise, relevant answers.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Make request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	// Check for rate limiting
	if resp.StatusCode == 429 {
		return "", ErrRateLimited
	}

	// Check for other errors
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Extract response text
	if len(openAIResp.Choices) == 0 {
		return "", ErrEmptyResponse
	}

	return openAIResp.Choices[0].Message.Content, nil
}
