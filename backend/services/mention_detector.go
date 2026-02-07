package services

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/Sneh16Shah/ai-visibility-tracker/db"
	"github.com/Sneh16Shah/ai-visibility-tracker/models"
)

// MentionDetector handles brand and competitor mention detection
type MentionDetector struct{}

// NewMentionDetector creates a new mention detector
func NewMentionDetector() *MentionDetector {
	return &MentionDetector{}
}

// DetectedMention represents a detected mention before storage
type DetectedMention struct {
	EntityName       string
	EntityType       string // "brand" or "competitor"
	Sentiment        string // "positive", "neutral", "negative"
	ContextSnippet   string
	Position         int
	IsRecommendation bool // True if explicitly recommended
	PositionRank     int  // 1=first, 2=second, 3+=later (within response)
}

// DetectMentions finds all brand and competitor mentions in AI response text
func (d *MentionDetector) DetectMentions(responseText string, brand *models.Brand) []DetectedMention {
	var mentions []DetectedMention

	// Normalize text for case-insensitive matching
	lowerText := strings.ToLower(responseText)

	// Detect brand mentions
	brandMentions := d.findEntityMentions(responseText, lowerText, brand.Name, "brand")
	mentions = append(mentions, brandMentions...)

	// Detect alias mentions
	for _, alias := range brand.Aliases {
		aliasMentions := d.findEntityMentions(responseText, lowerText, alias.Alias, "brand")
		mentions = append(mentions, aliasMentions...)
	}

	// Detect competitor mentions
	for _, competitor := range brand.Competitors {
		compMentions := d.findEntityMentions(responseText, lowerText, competitor.Name, "competitor")
		mentions = append(mentions, compMentions...)
	}

	// Sort mentions by position to assign position ranks
	sortMentionsByPosition(mentions)

	// Calculate position ranks per entity type
	brandRank := 0
	for i := range mentions {
		if mentions[i].EntityType == "brand" {
			brandRank++
			mentions[i].PositionRank = brandRank
		}

		// Analyze sentiment for each mention
		mentions[i].Sentiment = d.analyzeSentiment(mentions[i].ContextSnippet)

		// Check if this mention is an explicit recommendation
		mentions[i].IsRecommendation = d.isRecommendation(lowerText, mentions[i].EntityName, mentions[i].Position)
	}

	return mentions
}

// sortMentionsByPosition sorts mentions by their position in the text
func sortMentionsByPosition(mentions []DetectedMention) {
	for i := 0; i < len(mentions)-1; i++ {
		for j := i + 1; j < len(mentions); j++ {
			if mentions[j].Position < mentions[i].Position {
				mentions[i], mentions[j] = mentions[j], mentions[i]
			}
		}
	}
}

// Recommendation patterns - phrases that indicate explicit endorsement
var recommendationPatterns = []string{
	"i recommend",
	"i'd recommend",
	"we recommend",
	"i strongly recommend",
	"highly recommend",
	"my recommendation is",
	"is the best choice",
	"is the best option",
	"is my top pick",
	"is my top choice",
	"you should use",
	"you should go with",
	"go with",
	"i suggest",
	"i'd suggest",
	"the best option is",
	"the best choice is",
	"top pick",
	"first choice",
	"stands out as",
	"is ideal for",
	"is perfect for",
}

// isRecommendation checks if a mention is explicitly recommended
func (d *MentionDetector) isRecommendation(lowerText, entityName string, entityPosition int) bool {
	lowerEntity := strings.ToLower(entityName)

	for _, pattern := range recommendationPatterns {
		patternPos := strings.Index(lowerText, pattern)
		if patternPos == -1 {
			continue
		}

		// Check if entity name is within 150 characters of the recommendation phrase
		// This accounts for phrases like "For your use case, I recommend X because..."
		distance := abs(patternPos - entityPosition)
		if distance < 150 {
			// Additional check: entity should appear after the pattern or very close before
			// e.g., "I recommend Salesforce" or "Salesforce is my recommendation"
			entityPos := strings.Index(lowerText, lowerEntity)
			if entityPos >= patternPos-50 { // Entity can be up to 50 chars before pattern
				return true
			}
		}
	}

	return false
}

// abs returns absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func (d *MentionDetector) findEntityMentions(originalText, lowerText, entityName, entityType string) []DetectedMention {
	var mentions []DetectedMention

	lowerEntity := strings.ToLower(entityName)

	// Find all occurrences
	searchStart := 0
	for {
		pos := strings.Index(lowerText[searchStart:], lowerEntity)
		if pos == -1 {
			break
		}

		actualPos := searchStart + pos

		// Check for word boundaries (not partial word matches)
		if d.isWordBoundary(lowerText, actualPos, len(lowerEntity)) {
			// Extract context snippet (50 chars before and after)
			contextStart := max(0, actualPos-50)
			contextEnd := min(len(originalText), actualPos+len(entityName)+50)
			context := originalText[contextStart:contextEnd]

			// Add ellipsis if truncated
			if contextStart > 0 {
				context = "..." + context
			}
			if contextEnd < len(originalText) {
				context = context + "..."
			}

			mentions = append(mentions, DetectedMention{
				EntityName:     entityName,
				EntityType:     entityType,
				ContextSnippet: context,
				Position:       actualPos,
			})
		}

		searchStart = actualPos + len(lowerEntity)
	}

	return mentions
}

// isWordBoundary checks if the match is at word boundaries
func (d *MentionDetector) isWordBoundary(text string, pos, length int) bool {
	// Check character before
	if pos > 0 {
		prevChar := rune(text[pos-1])
		if unicode.IsLetter(prevChar) || unicode.IsDigit(prevChar) {
			return false
		}
	}

	// Check character after
	endPos := pos + length
	if endPos < len(text) {
		nextChar := rune(text[endPos])
		if unicode.IsLetter(nextChar) || unicode.IsDigit(nextChar) {
			return false
		}
	}

	return true
}

// Sentiment analysis words
var (
	positiveWords = []string{
		"best", "excellent", "great", "amazing", "outstanding", "fantastic",
		"superior", "recommended", "top", "leading", "preferred", "favorite",
		"powerful", "efficient", "reliable", "innovative", "impressive",
		"love", "perfect", "awesome", "brilliant", "exceptional", "superb",
		"highly recommended", "top-rated", "must-have", "game-changer",
	}

	negativeWords = []string{
		"worst", "terrible", "awful", "poor", "bad", "disappointing",
		"inferior", "avoid", "limited", "outdated", "slow", "expensive",
		"complicated", "confusing", "unreliable", "buggy", "frustrating",
		"hate", "horrible", "dreadful", "useless", "overpriced", "lacking",
		"not recommended", "stay away", "problems", "issues", "fails",
	}

	// Negation words that flip sentiment
	negationWords = []string{
		"not", "no", "never", "neither", "nobody", "nothing", "nowhere",
		"hardly", "barely", "doesn't", "don't", "didn't", "won't", "isn't",
		"aren't", "wasn't", "weren't", "hasn't", "haven't", "hadn't",
	}
)

// analyzeSentiment performs rule-based sentiment analysis on context
func (d *MentionDetector) analyzeSentiment(context string) string {
	lowerContext := strings.ToLower(context)

	positiveScore := 0
	negativeScore := 0

	// Check for positive words
	for _, word := range positiveWords {
		if strings.Contains(lowerContext, word) {
			// Check for negation nearby
			if d.hasNearbyNegation(lowerContext, word) {
				negativeScore++
			} else {
				positiveScore++
			}
		}
	}

	// Check for negative words
	for _, word := range negativeWords {
		if strings.Contains(lowerContext, word) {
			// Check for negation nearby (double negative = positive)
			if d.hasNearbyNegation(lowerContext, word) {
				positiveScore++
			} else {
				negativeScore++
			}
		}
	}

	// Determine sentiment
	if positiveScore > negativeScore && positiveScore > 0 {
		return "positive"
	} else if negativeScore > positiveScore && negativeScore > 0 {
		return "negative"
	}
	return "neutral"
}

// hasNearbyNegation checks if there's a negation word near the target word
func (d *MentionDetector) hasNearbyNegation(text, targetWord string) bool {
	// Look for negation within 5 words before the target
	targetPos := strings.Index(text, targetWord)
	if targetPos == -1 {
		return false
	}

	// Get text before target (up to 30 chars)
	searchStart := max(0, targetPos-30)
	beforeText := text[searchStart:targetPos]

	for _, neg := range negationWords {
		if strings.Contains(beforeText, neg) {
			return true
		}
	}

	return false
}

// StoreMentions saves detected mentions to the database
func (d *MentionDetector) StoreMentions(aiResponseID int, mentions []DetectedMention) ([]models.Mention, error) {
	repo := db.NewMentionRepository()
	var storedMentions []models.Mention

	for _, m := range mentions {
		mention, err := repo.Create(
			aiResponseID,
			m.EntityName,
			m.EntityType,
			m.Sentiment,
			m.ContextSnippet,
			m.Position,
			m.IsRecommendation,
			m.PositionRank,
		)
		if err != nil {
			return storedMentions, err
		}
		storedMentions = append(storedMentions, *mention)
	}

	return storedMentions, nil
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AnalyzeSentimentWithAI uses AI for more accurate sentiment (optional enhancement)
func (d *MentionDetector) AnalyzeSentimentWithAI(context string) string {
	// For now, use rule-based. Can be enhanced with AI later.
	return d.analyzeSentiment(context)
}

// ExtractKeyPhrases extracts key phrases around the mention
func (d *MentionDetector) ExtractKeyPhrases(text string) []string {
	// Simple regex to extract phrases
	phrasePattern := regexp.MustCompile(`[A-Za-z][a-z]+ [a-z]+ [a-z]+`)
	matches := phrasePattern.FindAllString(text, -1)
	return matches
}
