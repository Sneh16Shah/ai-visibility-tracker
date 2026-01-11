package controllers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/db"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// getJWTSecret returns the JWT secret key - reads from env at runtime
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}
	return []byte(secret)
}

// Claims for JWT token
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// SignupRequest represents signup request body
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest represents login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the response with token
type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// Signup creates a new user account
func Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	repo := db.NewUserRepository()

	// Check if email already exists
	if repo.EmailExists(req.Email) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Create user
	user, err := repo.Create(req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	// Generate token
	token, err := generateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User: gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// Login authenticates a user
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	repo := db.NewUserRepository()

	// Get user by email
	user, err := repo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Verify password
	if !repo.VerifyPassword(user, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate token
	token, err := generateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User: gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// GetMe returns the current authenticated user
func GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	repo := db.NewUserRepository()
	user, err := repo.GetByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
	})
}

// generateToken creates a JWT token
func generateToken(userID int, email string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return getJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// OptionalAuthMiddleware extracts user info if token present, but doesn't require it
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, use default user ID 1
			c.Set("userID", 1)
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return getJWTSecret(), nil
		})

		if err == nil && token.Valid {
			c.Set("userID", claims.UserID)
			c.Set("email", claims.Email)
		} else {
			c.Set("userID", 1) // Fallback to default user
		}
		c.Next()
	}
}
