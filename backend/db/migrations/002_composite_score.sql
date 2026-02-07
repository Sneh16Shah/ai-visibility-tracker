-- Migration: Add composite visibility score columns
-- Run this to add new columns for the weighted composite scoring system

USE ai_visibility_tracker;

-- Add new columns to mentions table for recommendation tracking
ALTER TABLE mentions
ADD COLUMN IF NOT EXISTS is_recommendation BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS position_rank INT DEFAULT 0;

-- Add new columns to metric_snapshots for component scores
ALTER TABLE metric_snapshots
ADD COLUMN IF NOT EXISTS normalized_mention_rate DECIMAL(7,4) DEFAULT 0,
ADD COLUMN IF NOT EXISTS weighted_position_score DECIMAL(7,4) DEFAULT 0,
ADD COLUMN IF NOT EXISTS recommendation_rate DECIMAL(7,4) DEFAULT 0,
ADD COLUMN IF NOT EXISTS relative_sentiment_index DECIMAL(7,4) DEFAULT 0,
ADD COLUMN IF NOT EXISTS confidence_score DECIMAL(7,4) DEFAULT 0,
ADD COLUMN IF NOT EXISTS confidence_level ENUM('high', 'medium', 'low') DEFAULT 'medium',
ADD COLUMN IF NOT EXISTS response_count INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS category_avg_sentiment DECIMAL(5,2) DEFAULT 0;

-- Create index for faster recommendation queries
CREATE INDEX IF NOT EXISTS idx_mentions_recommendation ON mentions(is_recommendation);
