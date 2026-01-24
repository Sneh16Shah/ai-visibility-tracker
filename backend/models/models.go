package models

import "time"

// User represents a user of the system
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Brand represents a brand being tracked
type Brand struct {
	ID                int          `json:"id"`
	UserID            int          `json:"user_id"`
	Name              string       `json:"name"`
	Industry          string       `json:"industry"`
	AlertThreshold    float64      `json:"alert_threshold"`    // Score below which to send alert
	ScheduleFrequency string       `json:"schedule_frequency"` // "disabled", "daily", "weekly"
	LastScheduledRun  time.Time    `json:"last_scheduled_run"`
	Aliases           []BrandAlias `json:"aliases,omitempty"`
	Competitors       []Competitor `json:"competitors,omitempty"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

// BrandAlias represents an alternative name for a brand
type BrandAlias struct {
	ID        int       `json:"id"`
	BrandID   int       `json:"brand_id"`
	Alias     string    `json:"alias"`
	CreatedAt time.Time `json:"created_at"`
}

// Competitor represents a competitor brand
type Competitor struct {
	ID        int       `json:"id"`
	BrandID   int       `json:"brand_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Prompt represents a prompt template
type Prompt struct {
	ID          int       `json:"id"`
	Category    string    `json:"category"`
	Template    string    `json:"template"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// AIResponse represents a response from an AI model
type AIResponse struct {
	ID           int       `json:"id"`
	BrandID      int       `json:"brand_id"`
	PromptID     int       `json:"prompt_id"`
	PromptText   string    `json:"prompt_text"`
	ResponseText string    `json:"response_text"`
	ModelName    string    `json:"model_name"`
	Mentions     []Mention `json:"mentions,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Mention represents a detected mention in an AI response
type Mention struct {
	ID             int       `json:"id"`
	AIResponseID   int       `json:"ai_response_id"`
	EntityName     string    `json:"entity_name"`
	EntityType     string    `json:"entity_type"` // "brand" or "competitor"
	Sentiment      string    `json:"sentiment"`   // "positive", "neutral", "negative"
	ContextSnippet string    `json:"context_snippet"`
	Position       int       `json:"position"`
	CreatedAt      time.Time `json:"created_at"`
}

// MetricSnapshot represents aggregated metrics at a point in time
type MetricSnapshot struct {
	ID              int       `json:"id"`
	BrandID         int       `json:"brand_id"`
	VisibilityScore float64   `json:"visibility_score"`
	CitationShare   float64   `json:"citation_share"`
	MentionCount    int       `json:"mention_count"`
	PositiveCount   int       `json:"positive_count"`
	NeutralCount    int       `json:"neutral_count"`
	NegativeCount   int       `json:"negative_count"`
	SnapshotDate    time.Time `json:"snapshot_date"`
	CreatedAt       time.Time `json:"created_at"`
}

// ============================================
// Request/Response DTOs
// ============================================

// CreateBrandRequest is the request body for creating a brand
type CreateBrandRequest struct {
	Name        string   `json:"name" binding:"required"`
	Industry    string   `json:"industry"`
	Aliases     []string `json:"aliases"`
	Competitors []string `json:"competitors"`
}

// UpdateBrandRequest is the request body for updating a brand
type UpdateBrandRequest struct {
	Name     string `json:"name"`
	Industry string `json:"industry"`
}

// AddAliasRequest is the request body for adding an alias
type AddAliasRequest struct {
	Alias string `json:"alias" binding:"required"`
}

// AddCompetitorRequest is the request body for adding a competitor
type AddCompetitorRequest struct {
	Name string `json:"name" binding:"required"`
}

// RunAnalysisRequest is the request body for running analysis
type RunAnalysisRequest struct {
	BrandID   int   `json:"brand_id" binding:"required"`
	PromptIDs []int `json:"prompt_ids"`
}

// DashboardData represents the data for the dashboard
type DashboardData struct {
	VisibilityScore   float64             `json:"visibility_score"`
	CitationShare     float64             `json:"citation_share"`
	TotalMentions     int                 `json:"total_mentions"`
	SentimentScore    float64             `json:"sentiment_score"`
	Trends            []MetricSnapshot    `json:"trends"`
	CitationBreakdown []CitationBreakdown `json:"citation_breakdown"`
	CompetitorData    []CompetitorMetrics `json:"competitor_data"`
}

// CitationBreakdown represents citation share by entity
type CitationBreakdown struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Color string  `json:"color"`
}

// CompetitorMetrics represents metrics for competitor comparison
type CompetitorMetrics struct {
	Name     string `json:"name"`
	Mentions int    `json:"mentions"`
	Positive int    `json:"positive"`
	Neutral  int    `json:"neutral"`
	Negative int    `json:"negative"`
}
