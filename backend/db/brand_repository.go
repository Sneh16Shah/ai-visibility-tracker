package db

import (
	"database/sql"
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/models"
)

// BrandRepository handles brand database operations
type BrandRepository struct {
	db *sql.DB
}

// NewBrandRepository creates a new brand repository
func NewBrandRepository() *BrandRepository {
	return &BrandRepository{db: DB}
}

// Create creates a new brand with aliases and competitors
func (r *BrandRepository) Create(userID int, req models.CreateBrandRequest) (*models.Brand, error) {
	// Insert brand
	result, err := r.db.Exec(
		"INSERT INTO brands (user_id, name, industry) VALUES (?, ?, ?)",
		userID, req.Name, req.Industry,
	)
	if err != nil {
		return nil, err
	}

	brandID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Insert aliases
	for _, alias := range req.Aliases {
		_, err = r.db.Exec(
			"INSERT INTO brand_aliases (brand_id, alias) VALUES (?, ?)",
			brandID, alias,
		)
		if err != nil {
			return nil, err
		}
	}

	// Insert competitors
	for _, comp := range req.Competitors {
		_, err = r.db.Exec(
			"INSERT INTO competitors (brand_id, name) VALUES (?, ?)",
			brandID, comp,
		)
		if err != nil {
			return nil, err
		}
	}

	return r.GetByID(int(brandID))
}

// GetByID retrieves a brand by ID with aliases and competitors
func (r *BrandRepository) GetByID(id int) (*models.Brand, error) {
	brand := &models.Brand{}
	err := r.db.QueryRow(
		"SELECT id, user_id, name, industry, created_at, updated_at FROM brands WHERE id = ?",
		id,
	).Scan(&brand.ID, &brand.UserID, &brand.Name, &brand.Industry, &brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get aliases
	aliasRows, err := r.db.Query("SELECT id, brand_id, alias, created_at FROM brand_aliases WHERE brand_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer aliasRows.Close()

	for aliasRows.Next() {
		var alias models.BrandAlias
		if err := aliasRows.Scan(&alias.ID, &alias.BrandID, &alias.Alias, &alias.CreatedAt); err != nil {
			return nil, err
		}
		brand.Aliases = append(brand.Aliases, alias)
	}

	// Get competitors
	compRows, err := r.db.Query("SELECT id, brand_id, name, created_at FROM competitors WHERE brand_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer compRows.Close()

	for compRows.Next() {
		var comp models.Competitor
		if err := compRows.Scan(&comp.ID, &comp.BrandID, &comp.Name, &comp.CreatedAt); err != nil {
			return nil, err
		}
		brand.Competitors = append(brand.Competitors, comp)
	}

	return brand, nil
}

// GetAll retrieves all brands for a user
func (r *BrandRepository) GetAll(userID int) ([]models.Brand, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, name, industry, created_at, updated_at FROM brands WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []models.Brand
	for rows.Next() {
		var brand models.Brand
		if err := rows.Scan(&brand.ID, &brand.UserID, &brand.Name, &brand.Industry, &brand.CreatedAt, &brand.UpdatedAt); err != nil {
			return nil, err
		}

		// Get aliases for this brand
		aliasRows, err := r.db.Query("SELECT id, brand_id, alias, created_at FROM brand_aliases WHERE brand_id = ?", brand.ID)
		if err != nil {
			return nil, err
		}
		for aliasRows.Next() {
			var alias models.BrandAlias
			if err := aliasRows.Scan(&alias.ID, &alias.BrandID, &alias.Alias, &alias.CreatedAt); err != nil {
				aliasRows.Close()
				return nil, err
			}
			brand.Aliases = append(brand.Aliases, alias)
		}
		aliasRows.Close()

		// Get competitors for this brand
		compRows, err := r.db.Query("SELECT id, brand_id, name, created_at FROM competitors WHERE brand_id = ?", brand.ID)
		if err != nil {
			return nil, err
		}
		for compRows.Next() {
			var comp models.Competitor
			if err := compRows.Scan(&comp.ID, &comp.BrandID, &comp.Name, &comp.CreatedAt); err != nil {
				compRows.Close()
				return nil, err
			}
			brand.Competitors = append(brand.Competitors, comp)
		}
		compRows.Close()

		brands = append(brands, brand)
	}

	return brands, nil
}

// GetAllBrands retrieves ALL brands (for scheduler/admin)
func (r *BrandRepository) GetAllBrands() ([]models.Brand, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, name, industry, COALESCE(alert_threshold, 0), COALESCE(schedule_frequency, ''), COALESCE(last_scheduled_run, '1970-01-01'), created_at, updated_at FROM brands",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []models.Brand
	for rows.Next() {
		var brand models.Brand
		if err := rows.Scan(&brand.ID, &brand.UserID, &brand.Name, &brand.Industry, &brand.AlertThreshold, &brand.ScheduleFrequency, &brand.LastScheduledRun, &brand.CreatedAt, &brand.UpdatedAt); err != nil {
			return nil, err
		}
		brands = append(brands, brand)
	}
	return brands, nil
}

// UpdateLastScheduledRun updates the last scheduled run time for a brand
func (r *BrandRepository) UpdateLastScheduledRun(brandID int, runTime time.Time) error {
	_, err := r.db.Exec(
		"UPDATE brands SET last_scheduled_run = ? WHERE id = ?",
		runTime, brandID,
	)
	return err
}

// UpdateAlertSettings updates alert threshold and schedule for a brand
func (r *BrandRepository) UpdateAlertSettings(brandID int, threshold float64, frequency string) error {
	_, err := r.db.Exec(
		"UPDATE brands SET alert_threshold = ?, schedule_frequency = ? WHERE id = ?",
		threshold, frequency, brandID,
	)
	return err
}

// Update updates a brand
func (r *BrandRepository) Update(id int, req models.UpdateBrandRequest) (*models.Brand, error) {
	_, err := r.db.Exec(
		"UPDATE brands SET name = ?, industry = ? WHERE id = ?",
		req.Name, req.Industry, id,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

// Delete deletes a brand and all related data (cascade delete)
func (r *BrandRepository) Delete(id int) error {
	// Delete in order to respect foreign key constraints:
	// 1. First delete mentions (references ai_responses)
	_, err := r.db.Exec(`
		DELETE FROM mentions 
		WHERE ai_response_id IN (SELECT id FROM ai_responses WHERE brand_id = ?)
	`, id)
	if err != nil {
		return err
	}

	// 2. Delete AI responses
	_, err = r.db.Exec("DELETE FROM ai_responses WHERE brand_id = ?", id)
	if err != nil {
		return err
	}

	// 3. Delete metric snapshots
	_, err = r.db.Exec("DELETE FROM metric_snapshots WHERE brand_id = ?", id)
	if err != nil {
		return err
	}

	// 4. Delete brand aliases
	_, err = r.db.Exec("DELETE FROM brand_aliases WHERE brand_id = ?", id)
	if err != nil {
		return err
	}

	// 5. Delete competitors
	_, err = r.db.Exec("DELETE FROM competitors WHERE brand_id = ?", id)
	if err != nil {
		return err
	}

	// 6. Finally delete the brand itself
	_, err = r.db.Exec("DELETE FROM brands WHERE id = ?", id)
	return err
}

// AddAlias adds an alias to a brand
func (r *BrandRepository) AddAlias(brandID int, alias string) (*models.BrandAlias, error) {
	result, err := r.db.Exec(
		"INSERT INTO brand_aliases (brand_id, alias) VALUES (?, ?)",
		brandID, alias,
	)
	if err != nil {
		return nil, err
	}

	aliasID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	brandAlias := &models.BrandAlias{}
	err = r.db.QueryRow(
		"SELECT id, brand_id, alias, created_at FROM brand_aliases WHERE id = ?",
		aliasID,
	).Scan(&brandAlias.ID, &brandAlias.BrandID, &brandAlias.Alias, &brandAlias.CreatedAt)

	return brandAlias, err
}

// RemoveAlias removes an alias
func (r *BrandRepository) RemoveAlias(aliasID int) error {
	_, err := r.db.Exec("DELETE FROM brand_aliases WHERE id = ?", aliasID)
	return err
}

// GetAliases gets all aliases for a brand
func (r *BrandRepository) GetAliases(brandID int) ([]models.BrandAlias, error) {
	rows, err := r.db.Query("SELECT id, brand_id, alias, created_at FROM brand_aliases WHERE brand_id = ?", brandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aliases []models.BrandAlias
	for rows.Next() {
		var alias models.BrandAlias
		if err := rows.Scan(&alias.ID, &alias.BrandID, &alias.Alias, &alias.CreatedAt); err != nil {
			return nil, err
		}
		aliases = append(aliases, alias)
	}
	return aliases, nil
}

// AddCompetitor adds a competitor to a brand
func (r *BrandRepository) AddCompetitor(brandID int, name string) (*models.Competitor, error) {
	result, err := r.db.Exec(
		"INSERT INTO competitors (brand_id, name) VALUES (?, ?)",
		brandID, name,
	)
	if err != nil {
		return nil, err
	}

	compID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	comp := &models.Competitor{}
	err = r.db.QueryRow(
		"SELECT id, brand_id, name, created_at FROM competitors WHERE id = ?",
		compID,
	).Scan(&comp.ID, &comp.BrandID, &comp.Name, &comp.CreatedAt)

	return comp, err
}

// RemoveCompetitor removes a competitor
func (r *BrandRepository) RemoveCompetitor(competitorID int) error {
	_, err := r.db.Exec("DELETE FROM competitors WHERE id = ?", competitorID)
	return err
}

// GetCompetitors gets all competitors for a brand
func (r *BrandRepository) GetCompetitors(brandID int) ([]models.Competitor, error) {
	rows, err := r.db.Query("SELECT id, brand_id, name, created_at FROM competitors WHERE brand_id = ?", brandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var competitors []models.Competitor
	for rows.Next() {
		var comp models.Competitor
		if err := rows.Scan(&comp.ID, &comp.BrandID, &comp.Name, &comp.CreatedAt); err != nil {
			return nil, err
		}
		competitors = append(competitors, comp)
	}
	return competitors, nil
}
