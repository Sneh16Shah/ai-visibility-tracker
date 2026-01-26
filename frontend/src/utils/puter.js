// Puter.js utility wrapper for Multi-AI Comparison
// Provides access to GPT, Claude, Gemini, Llama without API keys

// Available models for comparison (2026 latest - from Puter.js)
// Use puter.ai.listModels() to get full list
export const AI_MODELS = [
    { id: 'gpt-5', name: 'GPT-5', provider: 'OpenAI', color: '#10a37f' },
    { id: 'claude-opus-4-1', name: 'Claude Opus 4.1', provider: 'Anthropic', color: '#d4a574' },
    { id: 'gemini-2.5-pro', name: 'Gemini 2.5 Pro', provider: 'Google', color: '#4285f4' },
    { id: 'llama-4-maverick', name: 'Llama 4 Maverick', provider: 'Meta', color: '#0668e1' },
]

// Check if Puter is available
export function isPuterAvailable() {
    return typeof window !== 'undefined' && window.puter && window.puter.ai
}

// Query a single AI model
export async function queryModel(prompt, modelId) {
    if (!isPuterAvailable()) {
        throw new Error('Puter.js not available')
    }

    try {
        const response = await window.puter.ai.chat(prompt, { model: modelId })
        return {
            success: true,
            model: modelId,
            response: typeof response === 'string' ? response : response.message?.content || response.toString(),
            timestamp: new Date().toISOString()
        }
    } catch (error) {
        const errorMsg = error.message || error.toString() || 'Unknown error'
        console.error(`Error querying ${modelId}:`, error)

        // Detect balance/credit issues
        const isBalanceError = errorMsg.toLowerCase().includes('balance') ||
            errorMsg.toLowerCase().includes('credit') ||
            errorMsg.toLowerCase().includes('funding') ||
            errorMsg.toLowerCase().includes('upgrade')

        return {
            success: false,
            model: modelId,
            error: isBalanceError
                ? 'Puter credits exhausted. Try disabling Compare Mode to use backend API.'
                : errorMsg,
            isBalanceError,
            timestamp: new Date().toISOString()
        }
    }
}

// Query multiple models in parallel
export async function queryMultipleModels(prompt, modelIds = AI_MODELS.map(m => m.id)) {
    const queries = modelIds.map(modelId => queryModel(prompt, modelId))
    const results = await Promise.allSettled(queries)

    return results.map((result, index) => {
        if (result.status === 'fulfilled') {
            return result.value
        }
        return {
            success: false,
            model: modelIds[index],
            error: result.reason?.message || 'Query failed',
            timestamp: new Date().toISOString()
        }
    })
}

// Extract brand mentions from AI response
export function extractMentions(responseText, brandName, aliases = [], competitors = []) {
    // DEBUG: Log all inputs to identify toLowerCase issue
    console.log('[extractMentions] DEBUG - Inputs:', {
        responseText: typeof responseText,
        brandName: typeof brandName,
        brandNameValue: brandName,
        aliases: aliases,
        aliasTypes: aliases.map(a => ({ value: a, type: typeof a })),
        competitors: competitors,
        competitorTypes: competitors.map(c => ({ value: c, type: typeof c }))
    })

    const mentions = []

    // Defensive: ensure responseText is a string
    const text = typeof responseText === 'string' ? responseText : String(responseText || '')
    const textLower = text.toLowerCase()

    // Defensive: ensure brandName is a string
    const brand = typeof brandName === 'string' ? brandName : String(brandName || '')

    // Check for brand mentions
    const brandTerms = [brand, ...aliases].filter(Boolean).map(t => typeof t === 'string' ? t : (t?.alias || t?.name || String(t)))
    for (const term of brandTerms) {
        if (!term) continue
        if (textLower.includes(term.toLowerCase())) {
            mentions.push({
                entityName: brand,
                entityType: 'brand',
                sentiment: guessSentiment(text, term),
                found: true
            })
            break
        }
    }

    // Check for competitor mentions
    for (const competitor of competitors) {
        // Safely get competitor name as string
        const competitorName = typeof competitor === 'string' ? competitor : (competitor?.name || '')
        if (!competitorName) continue

        if (textLower.includes(competitorName.toLowerCase())) {
            mentions.push({
                entityName: competitorName,
                entityType: 'competitor',
                sentiment: guessSentiment(text, competitorName),
                found: true
            })
        }
    }

    return mentions
}

// Simple sentiment guess based on context
function guessSentiment(text, term) {
    // Defensive: ensure text and term are strings
    const safeText = typeof text === 'string' ? text : String(text || '')
    const safeTerm = typeof term === 'string' ? term : String(term || '')

    const textLower = safeText.toLowerCase()
    const termLower = safeTerm.toLowerCase()
    const termIndex = textLower.indexOf(termLower)

    if (termIndex === -1) return 'neutral'

    // Get surrounding context (100 chars before and after)
    const start = Math.max(0, termIndex - 100)
    const end = Math.min(safeText.length, termIndex + safeTerm.length + 100)
    const context = textLower.slice(start, end)

    const positiveWords = ['best', 'great', 'excellent', 'recommend', 'top', 'leading', 'popular', 'favorite', 'trusted', 'reliable']
    const negativeWords = ['worst', 'avoid', 'poor', 'bad', 'issue', 'problem', 'limited', 'lacks', 'expensive', 'outdated']

    let positiveScore = 0
    let negativeScore = 0

    for (const word of positiveWords) {
        if (context.includes(word)) positiveScore++
    }
    for (const word of negativeWords) {
        if (context.includes(word)) negativeScore++
    }

    if (positiveScore > negativeScore) return 'positive'
    if (negativeScore > positiveScore) return 'negative'
    return 'neutral'
}

// Calculate visibility score for a single response
export function calculateResponseScore(mentions, brandName) {
    const brandMention = mentions.find(m => m.entityType === 'brand')
    const competitorMentions = mentions.filter(m => m.entityType === 'competitor')

    if (!brandMention?.found) return 0

    // Base score: 50 for being mentioned
    let score = 50

    // Sentiment bonus: +25 positive, 0 neutral, -25 negative
    if (brandMention.sentiment === 'positive') score += 25
    else if (brandMention.sentiment === 'negative') score -= 25

    // Competition factor: fewer competitors = higher share
    const totalMentions = 1 + competitorMentions.length
    const citationShare = (1 / totalMentions) * 25
    score += citationShare

    return Math.min(100, Math.max(0, Math.round(score)))
}
