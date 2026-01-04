package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	Port        string
	Environment string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	AIProvider  string
	OpenAIKey   string
	GeminiKey   string
	GroqKey     string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBUser:      getEnv("DB_USER", "root"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "ai_visibility_tracker"),
		AIProvider:  getEnv("AI_PROVIDER", "gemini"),
		OpenAIKey:   getEnv("OPENAI_API_KEY", ""),
		GeminiKey:   getEnv("GEMINI_API_KEY", ""),
		GroqKey:     getEnv("GROQ_API_KEY", ""),
	}
}

// GetDSN returns MySQL Data Source Name
func (c *Config) GetDSN() string {
	return c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
