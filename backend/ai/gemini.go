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

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		model:   "gemini-3-flash-preview", // Fast and free
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GeminiRequest represents the request to Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents a content block
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of the content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Query sends a prompt to Gemini and returns the response
func (p *GeminiProvider) Query(ctx context.Context, prompt string) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("Gemini API key not configured")
	}

	// Build request
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request URL with API key
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error
	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	// Extract text from response
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// IsAvailable checks if the Gemini provider is configured
func (p *GeminiProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetModelName returns the model name
func (p *GeminiProvider) GetModelName() string {
	return "gemini-1.5-flash"
}
