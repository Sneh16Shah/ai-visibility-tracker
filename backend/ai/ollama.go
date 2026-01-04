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

// OllamaProvider implements the Provider interface for local Ollama
type OllamaProvider struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// OllamaRequest represents the request body for Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama2" // Default model
	}

	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for local LLM
		},
	}
}

// IsAvailable checks if Ollama is running
func (p *OllamaProvider) IsAvailable() bool {
	resp, err := p.httpClient.Get(p.baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// GetModelName returns the model name
func (p *OllamaProvider) GetModelName() string {
	return "ollama/" + p.model
}

// Query sends a prompt to Ollama and returns the response
func (p *OllamaProvider) Query(ctx context.Context, prompt string) (string, error) {
	// Build request
	reqBody := OllamaRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if ollamaResp.Response == "" {
		return "", ErrEmptyResponse
	}

	return ollamaResp.Response, nil
}
