package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/models"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{db: DB}
}

// Create creates a new user with hashed password
func (r *UserRepository) Create(email, password, name string) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := r.db.Exec(
		"INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)",
		email, string(hashedPassword), name,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(id))
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		"SELECT id, email, name, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var passwordHash string

	err := r.db.QueryRow(
		"SELECT id, email, password_hash, name, created_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &passwordHash, &user.Name, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Store password hash temporarily for verification
	user.PasswordHash = passwordHash
	return user, nil
}

// VerifyPassword checks if the password matches the hash
func (r *UserRepository) VerifyPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// EmailExists checks if an email is already registered
func (r *UserRepository) EmailExists(email string) bool {
	var count int
	r.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	return count > 0
}

// CreateDefaultUser creates a default demo user if none exists
func (r *UserRepository) CreateDefaultUser() error {
	// Check if any user exists
	var count int
	r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)

	if count == 0 {
		// Create default user
		_, err := r.Create("demo@example.com", "demo123", "Demo User")
		return err
	}
	return nil
}

// Update updates user info
func (r *UserRepository) Update(id int, name string) (*models.User, error) {
	_, err := r.db.Exec(
		"UPDATE users SET name = ?, updated_at = ? WHERE id = ?",
		name, time.Now(), id,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

// ErrUserNotFound is returned when user is not found
var ErrUserNotFound = errors.New("user not found")
