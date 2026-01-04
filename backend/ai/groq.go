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

// GroqProvider implements the Provider interface for Groq
type GroqProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewGroqProvider creates a new Groq provider
func NewGroqProvider(apiKey string) *GroqProvider {
	return &GroqProvider{
		apiKey:  apiKey,
		baseURL: "https://api.groq.com/openai/v1/chat/completions",
		model:   "llama-3.3-70b-versatile", // Fast and free
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GroqRequest represents the request to Groq API (OpenAI compatible)
type GroqRequest struct {
	Model    string        `json:"model"`
	Messages []GroqMessage `json:"messages"`
}

// GroqMessage represents a chat message
type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GroqResponse represents the response from Groq API
type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Query sends a prompt to Groq and returns the response
func (p *GroqProvider) Query(ctx context.Context, prompt string) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("Groq API key not configured")
	}

	// Build request
	reqBody := GroqRequest{
		Model: p.model,
		Messages: []GroqMessage{
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
	var groqResp GroqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error
	if groqResp.Error != nil {
		return "", fmt.Errorf("Groq API error: %s", groqResp.Error.Message)
	}

	// Extract text from response
	if len(groqResp.Choices) == 0 || groqResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no response from Groq")
	}

	return groqResp.Choices[0].Message.Content, nil
}

// IsAvailable checks if the Groq provider is configured
func (p *GroqProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetModelName returns the model name
func (p *GroqProvider) GetModelName() string {
	return "groq-llama-3.3-70b"
}
