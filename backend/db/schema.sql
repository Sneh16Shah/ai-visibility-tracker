-- AI Visibility Tracker - MySQL Schema
-- Run this to create the database and tables

CREATE DATABASE IF NOT EXISTS ai_visibility_tracker;
USE ai_visibility_tracker;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Brands table
CREATE TABLE IF NOT EXISTS brands (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    industry VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Brand aliases table
CREATE TABLE IF NOT EXISTS brand_aliases (
    id INT AUTO_INCREMENT PRIMARY KEY,
    brand_id INT NOT NULL,
    alias VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE CASCADE
);

-- Competitors table
CREATE TABLE IF NOT EXISTS competitors (
    id INT AUTO_INCREMENT PRIMARY KEY,
    brand_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE CASCADE
);

-- Prompt templates table
CREATE TABLE IF NOT EXISTS prompts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category VARCHAR(100) NOT NULL,
    template TEXT NOT NULL,
    description VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- AI responses table
CREATE TABLE IF NOT EXISTS ai_responses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    brand_id INT NOT NULL,
    prompt_id INT NOT NULL,
    prompt_text TEXT NOT NULL,
    response_text TEXT NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE CASCADE,
    FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
);

-- Mentions table
CREATE TABLE IF NOT EXISTS mentions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ai_response_id INT NOT NULL,
    entity_name VARCHAR(255) NOT NULL,
    entity_type ENUM('brand', 'competitor') NOT NULL,
    sentiment ENUM('positive', 'neutral', 'negative') DEFAULT 'neutral',
    context_snippet TEXT,
    position INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ai_response_id) REFERENCES ai_responses(id) ON DELETE CASCADE
);

-- Metric snapshots table
CREATE TABLE IF NOT EXISTS metric_snapshots (
    id INT AUTO_INCREMENT PRIMARY KEY,
    brand_id INT NOT NULL,
    visibility_score DECIMAL(5,2),
    citation_share DECIMAL(5,2),
    mention_count INT DEFAULT 0,
    positive_count INT DEFAULT 0,
    neutral_count INT DEFAULT 0,
    negative_count INT DEFAULT 0,
    snapshot_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE CASCADE
);

-- Insert default prompt templates
INSERT INTO prompts (category, template, description) VALUES
('Best Tools', 'What are the best {category} tools in 2024?', 'Find top tools in a category'),
('Alternatives', 'What are the best alternatives to {competitor}?', 'Find alternatives to a competitor'),
('Comparison', 'Compare {brand} vs {competitor} for {use_case}', 'Direct comparison between brands'),
('Beginner', 'What {category} tool should a beginner use?', 'Recommendations for beginners'),
('Reviews', 'What do people say about {brand}?', 'General sentiment and reviews');

-- Create indexes for better query performance
CREATE INDEX idx_brands_user ON brands(user_id);
CREATE INDEX idx_aliases_brand ON brand_aliases(brand_id);
CREATE INDEX idx_competitors_brand ON competitors(brand_id);
CREATE INDEX idx_responses_brand ON ai_responses(brand_id);
CREATE INDEX idx_mentions_response ON mentions(ai_response_id);
CREATE INDEX idx_metrics_brand_date ON metric_snapshots(brand_id, snapshot_date);
