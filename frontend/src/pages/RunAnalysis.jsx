import { useState, useEffect, useRef, useCallback } from 'react'
import * as api from '../api/client'

// Default prompt templates (used when API is unavailable)
const defaultPromptTemplates = [
    { id: 1, category: 'Best Tools', template: 'What are the best {category} tools available today?', selected: true },
    { id: 2, category: 'Alternatives', template: 'What are the best alternatives to {competitor}?', selected: true },
    { id: 3, category: 'Recommendations', template: 'Which {category} solution would you recommend for a business?', selected: true },
    { id: 4, category: 'Comparison', template: 'Compare {brand} vs {competitor} - which is better?', selected: false },
    { id: 5, category: 'Reviews', template: 'What do users say about {brand}? Is it worth using?', selected: true },
    { id: 6, category: 'Features', template: 'What are the key features of {brand}?', selected: false },
]

export default function RunAnalysis() {
    const [templates, setTemplates] = useState(defaultPromptTemplates)
    const [isRunning, setIsRunning] = useState(false)
    const [progress, setProgress] = useState(0)
    const [results, setResults] = useState([])
    const [error, setError] = useState(null)
    const [analysisStatus, setAnalysisStatus] = useState(null)
    const [cooldownSeconds, setCooldownSeconds] = useState(0)

    // Brand selection
    const [brands, setBrands] = useState([])
    const [selectedBrandId, setSelectedBrandId] = useState(null)

    // Ref to prevent double-clicks and React re-render issues
    const isRunningRef = useRef(false)
    const lastRunTime = useRef(0)

    // Minimum time between runs (client-side debounce)
    const MIN_RUN_INTERVAL = 3000 // 3 seconds

    // Fetch brands on mount
    useEffect(() => {
        const fetchBrands = async () => {
            try {
                const data = await api.getBrands()
                setBrands(data.brands || [])
                // Auto-select first brand if available
                if (data.brands && data.brands.length > 0) {
                    setSelectedBrandId(data.brands[0].id)
                }
            } catch (err) {
                console.log('Error fetching brands:', err)
            }
        }
        fetchBrands()
    }, [])

    // Fetch previous analysis results when brand changes
    useEffect(() => {
        if (!selectedBrandId) return

        const fetchPreviousResults = async () => {
            try {
                const data = await api.getAnalysisResults(selectedBrandId)
                if (data.results && data.results.length > 0) {
                    // Transform backend results to match our results format
                    setResults(data.results.map(r => ({
                        id: r.id,
                        prompt: r.prompt_text,
                        response: r.response_text,
                        model: r.model_name,
                        timestamp: r.created_at,
                        status: 'complete'
                    })))
                }
            } catch (err) {
                // No previous results or API error - that's fine
                console.log('No previous results:', err)
            }
        }

        fetchPreviousResults()
    }, [selectedBrandId])

    // Fetch analysis status on mount and periodically
    useEffect(() => {
        const fetchStatus = async () => {
            try {
                const status = await api.getAnalysisStatus()
                setAnalysisStatus(status)
            } catch (err) {
                // API might not be available
                console.log('Analysis status not available:', err)
            }
        }

        fetchStatus()
        const interval = setInterval(fetchStatus, 10000) // Check every 10 seconds

        return () => clearInterval(interval)
    }, [])

    // Cooldown timer
    useEffect(() => {
        if (cooldownSeconds > 0) {
            const timer = setTimeout(() => {
                setCooldownSeconds(cooldownSeconds - 1)
            }, 1000)
            return () => clearTimeout(timer)
        }
    }, [cooldownSeconds])

    const toggleTemplate = (id) => {
        if (isRunning) return // Don't allow changes while running
        setTemplates(templates.map(t =>
            t.id === id ? { ...t, selected: !t.selected } : t
        ))
    }

    // Debounced run analysis function
    const runAnalysis = useCallback(async () => {
        // Prevent double-clicks and rapid re-runs
        const now = Date.now()
        if (isRunningRef.current) {
            console.log('Analysis already in progress, ignoring click')
            return
        }
        if (now - lastRunTime.current < MIN_RUN_INTERVAL) {
            console.log('Too soon since last run, ignoring click')
            setError('Please wait a few seconds before running again')
            return
        }

        // Set running state immediately
        isRunningRef.current = true
        lastRunTime.current = now
        setIsRunning(true)
        setProgress(0)
        setResults([])
        setError(null)

        try {
            // Check if we can run analysis
            const status = await api.getAnalysisStatus()
            if (!status.can_run_analysis) {
                const waitTime = status.rate_limit_status?.seconds_until_reset || 60
                throw {
                    reason: `Rate limited. Please wait ${waitTime} seconds.`,
                    retry_after_sec: waitTime
                }
            }

            // Simulate initial progress
            setProgress(10)

            // Get selected prompt IDs
            const selectedPromptIds = templates.filter(t => t.selected).map(t => t.id)

            // Run the analysis (backend handles rate limiting)
            setProgress(30)
            const result = await api.runAnalysis(selectedBrandId, selectedPromptIds)

            setProgress(80)

            // Transform results for display
            if (result.responses && result.responses.length > 0) {
                const formattedResults = result.responses.map(r => ({
                    id: r.id,
                    prompt: r.prompt_text,
                    model: r.model_name,
                    timestamp: new Date(r.created_at).toLocaleString(),
                    response: r.response_text,
                    mentions: r.mentions || []
                }))
                setResults(formattedResults)
            }

            setProgress(100)

            // Show success message if there were errors
            if (result.errors && result.errors.length > 0) {
                setError(`Completed with warnings: ${result.errors.join(', ')}`)
            }

        } catch (err) {
            console.error('Analysis failed:', err)

            // Handle rate limiting
            if (err.status === 429 || err.retry_after_sec) {
                const waitTime = err.retry_after_sec || 60
                setCooldownSeconds(waitTime)
                setError(err.reason || `Rate limited. Please wait ${waitTime} seconds.`)
            } else if (err.status === 409) {
                setError('Analysis already in progress. Please wait.')
            } else if (err.status === 503) {
                setError('AI service not configured. Set OPENAI_API_KEY or use Ollama.')
            } else {
                setError(err.message || err.reason || 'Failed to run analysis. Using demo mode.')

                // Fall back to demo results
                setResults([
                    {
                        id: 1,
                        prompt: 'What are the best project management tools in 2024?',
                        model: 'Demo Mode',
                        timestamp: new Date().toLocaleString(),
                        response: 'Based on user reviews, the top project management tools include...',
                        mentions: [
                            { entity_name: 'Your Brand', sentiment: 'positive', context_snippet: '...is highly recommended for teams...' },
                            { entity_name: 'Competitor A', sentiment: 'neutral', context_snippet: '...is also a popular choice...' },
                        ],
                    },
                ])
            }
        } finally {
            setIsRunning(false)
            isRunningRef.current = false
        }
    }, [templates, selectedBrandId])

    const selectedCount = templates.filter(t => t.selected).length
    const selectedBrand = brands.find(b => b.id === selectedBrandId)
    const canRun = !isRunning && selectedCount > 0 && cooldownSeconds === 0 && selectedBrandId

    return (
        <div className="space-y-8">
            {/* Page Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold text-[var(--text)]">Run Analysis</h2>
                    <p className="text-[var(--text-muted)] mt-1">Execute AI prompts to analyze brand visibility</p>
                </div>
                <div className="flex items-center gap-4">
                    {/* Rate Limit Status Indicator */}
                    {analysisStatus && (
                        <div className="text-sm text-[var(--text-muted)]">
                            {analysisStatus.provider_available ? (
                                <span className="flex items-center gap-2">
                                    <span className="w-2 h-2 rounded-full bg-green-400"></span>
                                    AI Ready ({analysisStatus.rate_limit_status?.calls_this_minute || 0}/3 calls)
                                </span>
                            ) : (
                                <span className="flex items-center gap-2">
                                    <span className="w-2 h-2 rounded-full bg-amber-400"></span>
                                    Demo Mode
                                </span>
                            )}
                        </div>
                    )}

                    {/* Brand Selector */}
                    <select
                        value={selectedBrandId || ''}
                        onChange={(e) => setSelectedBrandId(Number(e.target.value))}
                        className="px-4 py-2 bg-[var(--surface)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)]"
                    >
                        {brands.length === 0 ? (
                            <option value="">No brands - add one first</option>
                        ) : (
                            brands.map(brand => (
                                <option key={brand.id} value={brand.id}>
                                    {brand.name}
                                </option>
                            ))
                        )}
                    </select>

                    <button
                        onClick={runAnalysis}
                        disabled={!canRun}
                        className={`px-6 py-3 rounded-lg font-medium transition-all duration-200 flex items-center gap-2 ${!canRun
                            ? 'bg-[var(--surface-light)] text-[var(--text-muted)] cursor-not-allowed'
                            : 'bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white shadow-lg shadow-indigo-500/25'
                            }`}
                    >
                        {isRunning ? (
                            <>
                                <span className="animate-spin">⏳</span>
                                <span>Running...</span>
                            </>
                        ) : cooldownSeconds > 0 ? (
                            <>
                                <span>⏱️</span>
                                <span>Wait {cooldownSeconds}s</span>
                            </>
                        ) : (
                            <>
                                <span>🚀</span>
                                <span>Run Analysis ({selectedCount} prompts)</span>
                            </>
                        )}
                    </button>
                </div>
            </div>

            {/* Error Message */}
            {error && (
                <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 text-red-400">
                    <div className="flex items-center gap-2">
                        <span>⚠️</span>
                        <span>{error}</span>
                    </div>
                </div>
            )}

            {/* Progress Bar */}
            {isRunning && (
                <div className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)]">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-[var(--text)]">Processing prompts...</span>
                        <span className="text-sm text-[var(--text-muted)]">{progress}%</span>
                    </div>
                    <div className="w-full bg-[var(--surface-light)] rounded-full h-3 overflow-hidden">
                        <div
                            className="bg-gradient-to-r from-indigo-500 to-purple-600 h-full rounded-full transition-all duration-300"
                            style={{ width: `${progress}%` }}
                        />
                    </div>
                </div>
            )}

            {/* Prompt Templates */}
            <div className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)]">
                <h3 className="text-lg font-semibold text-[var(--text)] mb-4">Prompt Templates</h3>
                <div className="space-y-3">
                    {templates.map((template) => (
                        <div
                            key={template.id}
                            onClick={() => toggleTemplate(template.id)}
                            className={`p-4 rounded-xl border cursor-pointer transition-all duration-200 ${template.selected
                                ? 'border-[var(--primary)] bg-[var(--primary)]/10'
                                : 'border-[var(--surface-light)] hover:border-[var(--surface-light)] hover:bg-[var(--surface-light)]/50'
                                }`}
                        >
                            <div className="flex items-center gap-3">
                                <div className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${template.selected
                                    ? 'border-[var(--primary)] bg-[var(--primary)]'
                                    : 'border-[var(--surface-light)]'
                                    }`}>
                                    {template.selected && <span className="text-white text-xs">✓</span>}
                                </div>
                                <div className="flex-1">
                                    <span className="inline-block px-2 py-0.5 bg-[var(--surface-light)] text-[var(--text-muted)] rounded text-xs mb-1">
                                        {template.category}
                                    </span>
                                    <p className="text-[var(--text)] font-mono text-sm">{template.template}</p>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            {/* Results */}
            {results.length > 0 && (
                <div className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)]">
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="text-lg font-semibold text-[var(--text)]">Analysis Results</h3>
                        <span className="text-sm text-[var(--text-muted)]">{results.length} responses analyzed</span>
                    </div>
                    <div className="space-y-4">
                        {results.map((result) => (
                            <div
                                key={result.id}
                                className="p-4 rounded-xl border border-[var(--surface-light)] bg-[var(--background)]"
                            >
                                <div className="flex items-start justify-between mb-3">
                                    <div>
                                        <p className="text-[var(--text)] font-medium">{result.prompt}</p>
                                        <div className="flex items-center gap-3 mt-1">
                                            <span className="text-xs text-[var(--text-muted)]">Model: {result.model}</span>
                                            <span className="text-xs text-[var(--text-muted)]">{result.timestamp}</span>
                                        </div>
                                    </div>
                                </div>

                                {/* AI Response */}
                                {result.response && (
                                    <div className="mb-3 p-3 rounded-lg bg-[var(--surface)] text-sm text-[var(--text-muted)]">
                                        {result.response.substring(0, 300)}
                                        {result.response.length > 300 && '...'}
                                    </div>
                                )}

                                {/* Mentions */}
                                {result.mentions && result.mentions.length > 0 && (
                                    <div className="space-y-2">
                                        <p className="text-xs text-[var(--text-muted)] font-medium">Detected Mentions:</p>
                                        {result.mentions.map((mention, i) => (
                                            <div key={i} className="flex items-center gap-3 p-2 rounded-lg bg-[var(--surface)]">
                                                <span className={`px-2 py-0.5 rounded text-xs font-medium ${mention.sentiment === 'positive' ? 'bg-green-500/20 text-green-400' :
                                                    mention.sentiment === 'negative' ? 'bg-red-500/20 text-red-400' :
                                                        'bg-amber-500/20 text-amber-400'
                                                    }`}>
                                                    {mention.sentiment}
                                                </span>
                                                <span className="text-[var(--text)] font-medium">{mention.entity_name || mention.brand}</span>
                                                <span className="text-[var(--text-muted)] text-sm italic">{mention.context_snippet || mention.context}</span>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    )
}
