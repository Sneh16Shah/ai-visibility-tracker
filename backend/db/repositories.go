package db

import (
	"database/sql"

	"github.com/Sneh16Shah/ai-visibility-tracker/models"
)

// PromptRepository handles prompt database operations
type PromptRepository struct {
	db *sql.DB
}

// NewPromptRepository creates a new prompt repository
func NewPromptRepository() *PromptRepository {
	return &PromptRepository{db: DB}
}

// GetAll retrieves all active prompts
func (r *PromptRepository) GetAll() ([]models.Prompt, error) {
	rows, err := r.db.Query(
		"SELECT id, category, template, description, is_active, created_at FROM prompts WHERE is_active = true",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []models.Prompt
	for rows.Next() {
		var prompt models.Prompt
		if err := rows.Scan(&prompt.ID, &prompt.Category, &prompt.Template, &prompt.Description, &prompt.IsActive, &prompt.CreatedAt); err != nil {
			return nil, err
		}
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

// GetByID retrieves a prompt by ID
func (r *PromptRepository) GetByID(id int) (*models.Prompt, error) {
	prompt := &models.Prompt{}
	err := r.db.QueryRow(
		"SELECT id, category, template, description, is_active, created_at FROM prompts WHERE id = ?",
		id,
	).Scan(&prompt.ID, &prompt.Category, &prompt.Template, &prompt.Description, &prompt.IsActive, &prompt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return prompt, nil
}

// Create creates a new prompt
func (r *PromptRepository) Create(category, template, description string) (*models.Prompt, error) {
	result, err := r.db.Exec(
		"INSERT INTO prompts (category, template, description) VALUES (?, ?, ?)",
		category, template, description,
	)
	if err != nil {
		return nil, err
	}

	promptID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(promptID))
}

// AIResponseRepository handles AI response database operations
type AIResponseRepository struct {
	db *sql.DB
}

// NewAIResponseRepository creates a new AI response repository
func NewAIResponseRepository() *AIResponseRepository {
	return &AIResponseRepository{db: DB}
}

// Create creates a new AI response
func (r *AIResponseRepository) Create(brandID, promptID int, promptText, responseText, modelName string) (*models.AIResponse, error) {
	result, err := r.db.Exec(
		"INSERT INTO ai_responses (brand_id, prompt_id, prompt_text, response_text, model_name) VALUES (?, ?, ?, ?, ?)",
		brandID, promptID, promptText, responseText, modelName,
	)
	if err != nil {
		return nil, err
	}

	responseID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(responseID))
}

// GetByID retrieves an AI response by ID
func (r *AIResponseRepository) GetByID(id int) (*models.AIResponse, error) {
	response := &models.AIResponse{}
	err := r.db.QueryRow(
		"SELECT id, brand_id, prompt_id, prompt_text, response_text, model_name, created_at FROM ai_responses WHERE id = ?",
		id,
	).Scan(&response.ID, &response.BrandID, &response.PromptID, &response.PromptText, &response.ResponseText, &response.ModelName, &response.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Get mentions for this response
	mentionRows, err := r.db.Query(
		"SELECT id, ai_response_id, entity_name, entity_type, sentiment, context_snippet, position, created_at FROM mentions WHERE ai_response_id = ?",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer mentionRows.Close()

	for mentionRows.Next() {
		var mention models.Mention
		if err := mentionRows.Scan(&mention.ID, &mention.AIResponseID, &mention.EntityName, &mention.EntityType, &mention.Sentiment, &mention.ContextSnippet, &mention.Position, &mention.CreatedAt); err != nil {
			return nil, err
		}
		response.Mentions = append(response.Mentions, mention)
	}

	return response, nil
}

// GetByBrandID retrieves all AI responses for a brand
func (r *AIResponseRepository) GetByBrandID(brandID int) ([]models.AIResponse, error) {
	rows, err := r.db.Query(
		"SELECT id, brand_id, prompt_id, prompt_text, response_text, model_name, created_at FROM ai_responses WHERE brand_id = ? ORDER BY created_at DESC",
		brandID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []models.AIResponse
	for rows.Next() {
		var response models.AIResponse
		if err := rows.Scan(&response.ID, &response.BrandID, &response.PromptID, &response.PromptText, &response.ResponseText, &response.ModelName, &response.CreatedAt); err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

// MentionRepository handles mention database operations
type MentionRepository struct {
	db *sql.DB
}

// NewMentionRepository creates a new mention repository
func NewMentionRepository() *MentionRepository {
	return &MentionRepository{db: DB}
}

// Create creates a new mention
func (r *MentionRepository) Create(aiResponseID int, entityName, entityType, sentiment, contextSnippet string, position int) (*models.Mention, error) {
	result, err := r.db.Exec(
		"INSERT INTO mentions (ai_response_id, entity_name, entity_type, sentiment, context_snippet, position) VALUES (?, ?, ?, ?, ?, ?)",
		aiResponseID, entityName, entityType, sentiment, contextSnippet, position,
	)
	if err != nil {
		return nil, err
	}

	mentionID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	mention := &models.Mention{}
	err = r.db.QueryRow(
		"SELECT id, ai_response_id, entity_name, entity_type, sentiment, context_snippet, position, created_at FROM mentions WHERE id = ?",
		mentionID,
	).Scan(&mention.ID, &mention.AIResponseID, &mention.EntityName, &mention.EntityType, &mention.Sentiment, &mention.ContextSnippet, &mention.Position, &mention.CreatedAt)

	return mention, err
}

// GetByResponseID gets all mentions for an AI response
func (r *MentionRepository) GetByResponseID(aiResponseID int) ([]models.Mention, error) {
	rows, err := r.db.Query(
		"SELECT id, ai_response_id, entity_name, entity_type, sentiment, context_snippet, position, created_at FROM mentions WHERE ai_response_id = ?",
		aiResponseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mentions []models.Mention
	for rows.Next() {
		var mention models.Mention
		if err := rows.Scan(&mention.ID, &mention.AIResponseID, &mention.EntityName, &mention.EntityType, &mention.Sentiment, &mention.ContextSnippet, &mention.Position, &mention.CreatedAt); err != nil {
			return nil, err
		}
		mentions = append(mentions, mention)
	}
	return mentions, nil
}

// MetricRepository handles metric database operations
type MetricRepository struct {
	db *sql.DB
}

// NewMetricRepository creates a new metric repository
func NewMetricRepository() *MetricRepository {
	return &MetricRepository{db: DB}
}

// Create creates a new metric snapshot
func (r *MetricRepository) Create(snapshot *models.MetricSnapshot) (*models.MetricSnapshot, error) {
	result, err := r.db.Exec(
		"INSERT INTO metric_snapshots (brand_id, visibility_score, citation_share, mention_count, positive_count, neutral_count, negative_count, snapshot_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		snapshot.BrandID, snapshot.VisibilityScore, snapshot.CitationShare, snapshot.MentionCount, snapshot.PositiveCount, snapshot.NeutralCount, snapshot.NegativeCount, snapshot.SnapshotDate,
	)
	if err != nil {
		return nil, err
	}

	snapshotID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(snapshotID))
}

// GetByID retrieves a metric snapshot by ID
func (r *MetricRepository) GetByID(id int) (*models.MetricSnapshot, error) {
	snapshot := &models.MetricSnapshot{}
	err := r.db.QueryRow(
		"SELECT id, brand_id, visibility_score, citation_share, mention_count, positive_count, neutral_count, negative_count, snapshot_date, created_at FROM metric_snapshots WHERE id = ?",
		id,
	).Scan(&snapshot.ID, &snapshot.BrandID, &snapshot.VisibilityScore, &snapshot.CitationShare, &snapshot.MentionCount, &snapshot.PositiveCount, &snapshot.NeutralCount, &snapshot.NegativeCount, &snapshot.SnapshotDate, &snapshot.CreatedAt)
	return snapshot, err
}

// GetLatestByBrandID retrieves the latest metric snapshot for a brand
func (r *MetricRepository) GetLatestByBrandID(brandID int) (*models.MetricSnapshot, error) {
	snapshot := &models.MetricSnapshot{}
	err := r.db.QueryRow(
		"SELECT id, brand_id, visibility_score, citation_share, mention_count, positive_count, neutral_count, negative_count, snapshot_date, created_at FROM metric_snapshots WHERE brand_id = ? ORDER BY snapshot_date DESC LIMIT 1",
		brandID,
	).Scan(&snapshot.ID, &snapshot.BrandID, &snapshot.VisibilityScore, &snapshot.CitationShare, &snapshot.MentionCount, &snapshot.PositiveCount, &snapshot.NeutralCount, &snapshot.NegativeCount, &snapshot.SnapshotDate, &snapshot.CreatedAt)
	return snapshot, err
}

// GetTrendsByBrandID retrieves metric trends for a brand (last 7 days)
func (r *MetricRepository) GetTrendsByBrandID(brandID int, days int) ([]models.MetricSnapshot, error) {
	rows, err := r.db.Query(
		"SELECT id, brand_id, visibility_score, citation_share, mention_count, positive_count, neutral_count, negative_count, snapshot_date, created_at FROM metric_snapshots WHERE brand_id = ? ORDER BY snapshot_date DESC LIMIT ?",
		brandID, days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.MetricSnapshot
	for rows.Next() {
		var snapshot models.MetricSnapshot
		if err := rows.Scan(&snapshot.ID, &snapshot.BrandID, &snapshot.VisibilityScore, &snapshot.CitationShare, &snapshot.MentionCount, &snapshot.PositiveCount, &snapshot.NeutralCount, &snapshot.NegativeCount, &snapshot.SnapshotDate, &snapshot.CreatedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}
