import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import * as api from '../api/client'
import { AI_MODELS } from '../utils/models'

// Default prompt templates (used when API is unavailable)
const defaultPromptTemplates = [
    { id: 1, category: 'Best Tools', template: 'What are the best {category} tools available today?', selected: true },
    { id: 2, category: 'Alternatives', template: 'What are the best alternatives to {competitor}?', selected: true },
    { id: 3, category: 'Recommendations', template: 'Which {category} solution would you recommend for a business?', selected: true },
    { id: 4, category: 'Comparison', template: 'Compare {brand} vs {competitor} - which is better?', selected: false },
    { id: 5, category: 'Reviews', template: 'What do users say about {brand}? Is it worth using?', selected: true },
    { id: 6, category: 'Features', template: 'What are the key features of {brand}?', selected: false },
]

// Default categories - prompts NOT in this list are user-created and can be edited/deleted
const DEFAULT_CATEGORIES = ['Best Tools', 'Alternatives', 'Recommendations', 'Comparison', 'Reviews', 'Features']
const isCustomPrompt = (prompt) => !DEFAULT_CATEGORIES.includes(prompt.category)

export default function RunAnalysis() {
    const navigate = useNavigate()
    const [searchParams] = useSearchParams()

    const [templates, setTemplates] = useState(defaultPromptTemplates)
    const [isRunning, setIsRunning] = useState(false)
    const [progress, setProgress] = useState(0)
    const [results, setResults] = useState([])
    const [error, setError] = useState(null)
    const [analysisStatus, setAnalysisStatus] = useState(null)
    const [cooldownSeconds, setCooldownSeconds] = useState(0)
    const [showCompletionModal, setShowCompletionModal] = useState(false)
    const [showAddPrompt, setShowAddPrompt] = useState(false)
    const [newPromptTemplate, setNewPromptTemplate] = useState('')
    const [newPromptCategory, setNewPromptCategory] = useState('Custom')

    // Inline editing and delete modal state
    const [editingPromptId, setEditingPromptId] = useState(null)
    const [editingText, setEditingText] = useState('')
    const [deleteModalPrompt, setDeleteModalPrompt] = useState(null)

    // Brand selection
    const [brands, setBrands] = useState([])
    const [selectedBrandId, setSelectedBrandId] = useState(null)

    // Compare Mode (Multi-AI) - Uses OpenRouter backend
    const [compareMode, setCompareMode] = useState(false)
    const [selectedModels, setSelectedModels] = useState([
        'google/gemma-3-27b-it:free',
        'meta-llama/llama-3.3-70b-instruct:free',
        'qwen/qwen3-coder:free',
        'tngtech/deepseek-r1t2-chimera:free',
        'groq'
    ])
    const [compareResults, setCompareResults] = useState([])

    // Expand/collapse state for results
    const [expandedResults, setExpandedResults] = useState({})

    // Ref to prevent double-clicks and React re-render issues
    const isRunningRef = useRef(false)
    const lastRunTime = useRef(0)
    const abortControllerRef = useRef(null)

    // Minimum time between runs (client-side debounce)
    const MIN_RUN_INTERVAL = 3000 // 3 seconds

    // Cancel analysis function
    const cancelAnalysis = useCallback(() => {
        if (abortControllerRef.current) {
            abortControllerRef.current.abort()
            abortControllerRef.current = null
        }
        setIsRunning(false)
        isRunningRef.current = false
        setProgress(0)
        setError('Analysis cancelled')
    }, [])

    // Fetch brands on mount
    useEffect(() => {
        const fetchBrands = async () => {
            try {
                const data = await api.getBrands()
                setBrands(data.brands || [])

                // Check for brand_id in URL params first
                const urlBrandId = searchParams.get('brand_id')
                if (urlBrandId && data.brands?.some(b => b.id === parseInt(urlBrandId))) {
                    setSelectedBrandId(parseInt(urlBrandId))
                } else if (data.brands && data.brands.length > 0) {
                    // Fall back to first brand
                    setSelectedBrandId(data.brands[0].id)
                }
            } catch (err) {
                console.log('Error fetching brands:', err)
            }
        }
        fetchBrands()
    }, [searchParams])

    // Fetch prompts from API on mount (includes custom questions)
    useEffect(() => {
        const fetchPrompts = async () => {
            try {
                const data = await api.getPrompts()
                if (data.prompts && data.prompts.length > 0) {
                    // Merge API prompts with selection state
                    setTemplates(data.prompts.map(p => ({
                        id: p.id,
                        category: p.category,
                        template: p.template,
                        description: p.description,
                        selected: p.category !== 'Features' // Default selection
                    })))
                }
            } catch (err) {
                console.log('Could not fetch prompts, using defaults:', err)
                // Keep default templates if API fails
            }
        }
        fetchPrompts()
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

    // Get selected brand info
    const getSelectedBrand = () => {
        return brands.find(b => b.id === selectedBrandId)
    }

    // Run Compare Mode analysis (multi-model via OpenRouter backend)
    const runCompareAnalysis = useCallback(async () => {
        if (selectedModels.length === 0) {
            setError('Please select at least one AI model to compare')
            return
        }

        const brand = getSelectedBrand()
        if (!brand) {
            setError('Please select a brand first')
            return
        }

        isRunningRef.current = true
        lastRunTime.current = Date.now()
        setIsRunning(true)
        setProgress(0)
        setCompareResults([])
        setError(null)

        try {
            // Get selected prompt IDs
            const selectedPromptIds = templates.filter(t => t.selected).map(t => t.id)
            if (selectedPromptIds.length === 0) {
                throw new Error('Please select at least one prompt')
            }

            setProgress(20)

            // Call backend API for multi-model comparison
            const result = await api.runCompareModels(selectedBrandId, selectedPromptIds, selectedModels)

            setProgress(80)

            if (!result.success && result.results?.length === 0) {
                throw new Error(result.message || 'Comparison failed')
            }

            // Transform backend results to frontend format
            const allModelResults = result.results.map(r => {
                const modelInfo = AI_MODELS.find(m => m.id === r.model_id)
                return {
                    id: `${r.model_id}-${Date.now()}-${Math.random()}`,
                    model: r.model_name || modelInfo?.name || r.model_id,
                    modelId: r.model_id,
                    provider: r.provider || modelInfo?.provider || 'Unknown',
                    color: r.color || modelInfo?.color || '#888888',
                    prompt: r.prompt_text,
                    response: r.response,
                    mentions: r.mentions || [],
                    score: r.score || 0,
                    error: r.error,
                    timestamp: r.timestamp
                }
            })

            setCompareResults(allModelResults)
            setProgress(100)
            setShowCompletionModal(true)

            // Save per-model scores to localStorage for Dashboard
            const modelScores = {}
            allModelResults.forEach(r => {
                if (r.error) return // Skip failed results
                if (!modelScores[r.model]) {
                    modelScores[r.model] = { total: 0, count: 0, color: r.color, modelId: r.modelId }
                }
                modelScores[r.model].total += r.score
                modelScores[r.model].count++
            })
            const scoreSummary = Object.entries(modelScores)
                .filter(([, data]) => data.count > 0)
                .map(([model, data]) => ({
                    model,
                    modelId: data.modelId,
                    color: data.color,
                    score: Math.round(data.total / data.count)
                }))
            // Store with brand ID so each brand has its own compare data
            const stored = JSON.parse(localStorage.getItem('compareModelScores') || '{}')
            stored[brand.id] = { scores: scoreSummary, timestamp: new Date().toISOString() }
            localStorage.setItem('compareModelScores', JSON.stringify(stored))

        } catch (err) {
            console.error('Compare analysis failed:', err)
            if (err.status === 503) {
                setError('Compare Models requires OPENROUTER_API_KEY. Please configure it in your .env file.')
            } else {
                setError(err.message || 'Comparison failed')
            }
        } finally {
            setIsRunning(false)
            isRunningRef.current = false
        }
    }, [selectedModels, templates, selectedBrandId])

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

            // Show completion modal on success
            if (!result.errors || result.errors.length === 0) {
                setShowCompletionModal(true)
            } else {
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
                                    AI Ready ({analysisStatus.rate_limit_status?.calls_this_minute || 0}/{analysisStatus.rate_limit_status?.max_calls_per_minute || 10} calls)
                                </span>
                            ) : (
                                <span className="flex items-center gap-2">
                                    <span className="w-2 h-2 rounded-full bg-amber-400"></span>
                                    Demo Mode
                                </span>
                            )}
                        </div>
                    )}

                    {/* Compare Models Toggle */}
                    <button
                        onClick={() => setCompareMode(!compareMode)}
                        className={`px-4 py-2 rounded-lg font-medium transition-all duration-200 flex items-center gap-2 ${compareMode
                            ? 'bg-purple-500/20 text-purple-400 border border-purple-500/50'
                            : 'bg-[var(--surface)] text-[var(--text-muted)] border border-[var(--surface-light)] hover:border-[var(--primary)]'
                            }`}
                    >
                        <span>{compareMode ? '✓' : ''}</span>
                        <span>Compare Models</span>
                    </button>

                    {/* Brand Selector */}
                    <select
                        value={selectedBrandId || ''}
                        onChange={(e) => setSelectedBrandId(Number(e.target.value))}
                        className="select"
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
                        onClick={compareMode ? runCompareAnalysis : runAnalysis}
                        disabled={!canRun || (compareMode && selectedModels.length === 0)}
                        className={`btn ${(!canRun || (compareMode && selectedModels.length === 0))
                            ? 'btn-secondary opacity-50 cursor-not-allowed'
                            : compareMode
                                ? 'btn-compare'
                                : 'btn-primary'
                            } px-6 py-3`}
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

                    {/* Cancel button - show only when running */}
                    {isRunning && (
                        <button
                            onClick={cancelAnalysis}
                            className="btn px-4 py-3 border border-[var(--error)]/50 text-[var(--error)] hover:bg-[var(--error)]/10"
                        >
                            <span>✕</span>
                            <span>Cancel</span>
                        </button>
                    )}
                </div>
            </div>

            {/* Error Message */}
            {error && (
                <div className="alert alert-error">
                    <span>⚠️</span>
                    <span>{error}</span>
                </div>
            )}

            {/* Progress Bar */}
            {isRunning && (
                <div className="card">
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-[var(--text)]">Processing prompts...</span>
                        <span className="text-sm text-[var(--text-muted)]">{progress}%</span>
                    </div>
                    <div className="progress-track h-3">
                        <div
                            className="progress-bar"
                            style={{ width: `${progress}%` }}
                        />
                    </div>
                </div>
            )}

            {/* AI Model Selection (Compare Mode) */}
            {compareMode && (
                <div className="card card-accent-purple">
                    <h3 className="text-lg font-semibold text-[var(--text)] mb-4 flex items-center gap-2">
                        <span>🤖</span> Select AI Models to Compare
                    </h3>
                    <div className="flex flex-wrap gap-3">
                        {AI_MODELS.map((model) => {
                            const isSelected = selectedModels.includes(model.id)
                            return (
                                <button
                                    key={model.id}
                                    onClick={() => {
                                        if (isSelected) {
                                            setSelectedModels(prev => prev.filter(id => id !== model.id))
                                        } else {
                                            setSelectedModels(prev => [...prev, model.id])
                                        }
                                    }}
                                    className={`btn ${isSelected
                                        ? 'text-white shadow-lg'
                                        : 'btn-secondary'
                                        }`}
                                    style={{
                                        backgroundColor: isSelected ? model.color : undefined,
                                    }}
                                >
                                    <span>{isSelected ? '✓' : ''}</span>
                                    <span>{model.name}</span>
                                    <span className="text-xs opacity-70">({model.provider})</span>
                                </button>
                            )
                        })}
                    </div>
                    {selectedModels.length === 0 && (
                        <p className="text-[var(--warning)] text-sm mt-3">⚠️ Select at least one model to run comparison</p>
                    )}
                </div>
            )}

            {/* Prompt Templates */}
            <div className="card">
                <h3 className="text-lg font-semibold text-[var(--text)] mb-4">Prompt Templates</h3>
                <div className="space-y-3">
                    {templates.map((template) => (
                        <div
                            key={template.id}
                            onClick={() => editingPromptId !== template.id && toggleTemplate(template.id)}
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
                                    {/* Inline editing for Custom prompts */}
                                    {editingPromptId === template.id ? (
                                        <div className="flex gap-2 mt-1" onClick={(e) => e.stopPropagation()}>
                                            <input
                                                type="text"
                                                value={editingText}
                                                onChange={(e) => setEditingText(e.target.value)}
                                                className="flex-1 px-3 py-2 bg-[var(--background)] border border-[var(--primary)] rounded-lg text-[var(--text)] text-sm focus:outline-none"
                                                autoFocus
                                                onKeyDown={(e) => {
                                                    if (e.key === 'Enter') {
                                                        e.preventDefault();
                                                        (async () => {
                                                            try {
                                                                await api.updatePrompt(template.id, 'Custom', editingText, '');
                                                                setTemplates(prev => prev.map(t =>
                                                                    t.id === template.id ? { ...t, template: editingText } : t
                                                                ));
                                                                setEditingPromptId(null);
                                                            } catch (err) {
                                                                console.error('Failed to update:', err);
                                                            }
                                                        })();
                                                    } else if (e.key === 'Escape') {
                                                        setEditingPromptId(null);
                                                    }
                                                }}
                                            />
                                            <button
                                                onClick={async () => {
                                                    try {
                                                        await api.updatePrompt(template.id, 'Custom', editingText, '');
                                                        setTemplates(prev => prev.map(t =>
                                                            t.id === template.id ? { ...t, template: editingText } : t
                                                        ));
                                                        setEditingPromptId(null);
                                                    } catch (err) {
                                                        console.error('Failed to update:', err);
                                                    }
                                                }}
                                                className="px-3 py-2 bg-green-500/20 text-green-400 rounded-lg hover:bg-green-500/30"
                                            >
                                                ✓
                                            </button>
                                            <button
                                                onClick={() => setEditingPromptId(null)}
                                                className="px-3 py-2 bg-red-500/20 text-red-400 rounded-lg hover:bg-red-500/30"
                                            >
                                                ✕
                                            </button>
                                        </div>
                                    ) : (
                                        <p className="text-[var(--text)] font-mono text-sm">{template.template}</p>
                                    )}
                                </div>
                                {/* Edit/Delete for Custom (user-created) prompts */}
                                {isCustomPrompt(template) && editingPromptId !== template.id && (
                                    <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
                                        <button
                                            onClick={() => {
                                                setEditingPromptId(template.id);
                                                setEditingText(template.template);
                                            }}
                                            className="p-1.5 text-[var(--text-muted)] hover:text-blue-400 transition-colors"
                                            title="Edit"
                                        >
                                            ✏️
                                        </button>
                                        <button
                                            onClick={() => setDeleteModalPrompt(template)}
                                            className="p-1.5 text-[var(--text-muted)] hover:text-red-400 transition-colors"
                                            title="Delete"
                                        >
                                            🗑️
                                        </button>
                                    </div>
                                )}
                            </div>
                        </div>
                    ))}
                </div>

                {/* Add Custom Question */}
                {!showAddPrompt ? (
                    <button
                        onClick={() => setShowAddPrompt(true)}
                        className="mt-4 w-full py-3 border-2 border-dashed border-[var(--surface-light)] rounded-xl text-[var(--text-muted)] hover:border-[var(--primary)] hover:text-[var(--primary)] transition-all duration-200 flex items-center justify-center gap-2"
                    >
                        <span>+</span>
                        <span>Add Custom Question</span>
                    </button>
                ) : (
                    <div className="mt-4 p-4 border border-[var(--surface-light)] rounded-xl bg-[var(--background)]">
                        <div className="flex gap-3 mb-3">
                            <input
                                type="text"
                                value={newPromptCategory}
                                onChange={(e) => setNewPromptCategory(e.target.value)}
                                placeholder="Category"
                                className="w-32 px-3 py-2 bg-[var(--surface)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] text-sm focus:outline-none focus:border-[var(--primary)]"
                            />
                            <input
                                type="text"
                                value={newPromptTemplate}
                                onChange={(e) => setNewPromptTemplate(e.target.value)}
                                placeholder="Enter your question... Use {brand}, {competitor}, {category} for placeholders"
                                className="flex-1 px-3 py-2 bg-[var(--surface)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] text-sm focus:outline-none focus:border-[var(--primary)]"
                            />
                        </div>
                        <div className="flex gap-2 justify-end">
                            <button
                                onClick={() => {
                                    setShowAddPrompt(false)
                                    setNewPromptTemplate('')
                                    setNewPromptCategory('Custom')
                                }}
                                className="px-4 py-2 text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={async () => {
                                    if (!newPromptTemplate.trim()) return
                                    try {
                                        const created = await api.createPrompt(newPromptCategory, newPromptTemplate)
                                        setTemplates(prev => [...prev, { ...created, selected: true }])
                                        setShowAddPrompt(false)
                                        setNewPromptTemplate('')
                                        setNewPromptCategory('Custom')
                                    } catch (err) {
                                        console.error('Failed to create prompt:', err)
                                        setError('Failed to create custom prompt')
                                    }
                                }}
                                className="px-4 py-2 bg-[var(--primary)] text-white rounded-lg hover:opacity-90 transition-opacity"
                            >
                                Add Question
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* Compare Results (Multi-Model) */}
            {compareMode && compareResults.length > 0 && (
                <div className="card card-accent-purple">
                    <div className="flex items-center justify-between mb-6">
                        <h3 className="text-lg font-semibold text-[var(--text)] flex items-center gap-2">
                            <span>📊</span> Model Comparison Results
                        </h3>
                        <span className="text-sm text-[var(--text-muted)]">
                            {compareResults.length} responses from {selectedModels.length} models
                        </span>
                    </div>

                    {/* Score Summary by Model */}
                    <div className="mb-6">
                        <h4 className="text-sm font-medium text-[var(--text-muted)] mb-3">Visibility Score by Model</h4>
                        <div className="space-y-2">
                            {(() => {
                                // Calculate average score per model
                                const modelScores = {}
                                compareResults.forEach(r => {
                                    if (!modelScores[r.model]) {
                                        modelScores[r.model] = { total: 0, count: 0, color: r.color }
                                    }
                                    modelScores[r.model].total += r.score
                                    modelScores[r.model].count++
                                })
                                return Object.entries(modelScores).map(([model, data]) => {
                                    const avgScore = Math.round(data.total / data.count)
                                    return (
                                        <div key={model} className="flex items-center gap-3">
                                            <span className="w-28 text-sm text-[var(--text)]">{model}</span>
                                            <div className="flex-1 bg-[var(--surface-light)] rounded-full h-4 overflow-hidden">
                                                <div
                                                    className="h-full rounded-full transition-all duration-500"
                                                    style={{
                                                        width: `${avgScore}%`,
                                                        backgroundColor: data.color
                                                    }}
                                                />
                                            </div>
                                            <span className="w-12 text-sm font-mono text-[var(--text)]">{avgScore}</span>
                                        </div>
                                    )
                                })
                            })()}
                        </div>
                    </div>

                    {/* Detailed Results Grid */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {compareResults.map((result) => (
                            <div
                                key={result.id}
                                className="p-4 rounded-xl border bg-[var(--background)]"
                                style={{ borderColor: result.color + '50' }}
                            >
                                <div className="flex items-center justify-between mb-2">
                                    <div className="flex items-center gap-2">
                                        <div
                                            className="w-3 h-3 rounded-full"
                                            style={{ backgroundColor: result.color }}
                                        />
                                        <span className="font-medium text-[var(--text)]">{result.model}</span>
                                        <span className="text-xs text-[var(--text-muted)]">({result.provider})</span>
                                    </div>
                                    <span
                                        className="text-sm font-mono px-2 py-0.5 rounded"
                                        style={{
                                            backgroundColor: result.color + '20',
                                            color: result.color
                                        }}
                                    >
                                        Score: {result.score}
                                    </span>
                                </div>
                                <p className="text-xs text-[var(--text-muted)] mb-2 font-mono">{result.prompt}</p>
                                {result.error ? (
                                    <p className="text-red-400 text-sm">Error: {result.error}</p>
                                ) : (
                                    <>
                                        {/* Expand/Collapse Toggle */}
                                        <div className="relative">
                                            <button
                                                onClick={() => setExpandedResults(prev => ({ ...prev, [result.id]: !prev[result.id] }))}
                                                className="text-xs text-[var(--primary)] hover:underline mb-2 flex items-center gap-1"
                                            >
                                                {expandedResults[result.id] ? '🔽 Collapse Response' : '🔼 Expand Full Response'}
                                            </button>
                                            <div className={`text-sm text-[var(--text)] ${expandedResults[result.id] ? 'max-h-96 overflow-y-auto' : 'max-h-20 overflow-hidden'} transition-all duration-200 bg-[var(--surface)] p-3 rounded-lg`}>
                                                {expandedResults[result.id] ? (
                                                    // Full formatted response
                                                    <div className="space-y-2">
                                                        {result.response?.split('\n').map((line, li) => {
                                                            // Strip all ** and * markers
                                                            const cleanText = (text) => text.replace(/\*\*/g, '').replace(/\*/g, '');

                                                            // Skip separator lines
                                                            if (line.trim() === '---' || line.trim() === '***') {
                                                                return <hr key={li} className="border-[var(--surface-light)] my-3" />;
                                                            }
                                                            // Markdown headings (### Title)
                                                            if (line.startsWith('###')) {
                                                                return <p key={li} className="font-bold text-base text-[var(--primary)] mt-3">{cleanText(line.replace(/^###\s*/, ''))}</p>;
                                                            }
                                                            // Bold headings (lines starting with **)
                                                            if (line.startsWith('**')) {
                                                                return <p key={li} className="font-bold text-[var(--primary)] mt-4">{cleanText(line)}</p>;
                                                            }
                                                            // Regular numbered headings (1. Title:)
                                                            if (/^\d+\.\s/.test(line)) {
                                                                return <p key={li} className="font-semibold text-[var(--primary)] mt-3">{cleanText(line)}</p>;
                                                            }
                                                            // Bullet points (check for - or single * but not **)
                                                            if (line.startsWith('- ') || (line.startsWith('* ') && !line.startsWith('**'))) {
                                                                return <p key={li} className="pl-4 text-[var(--text)]">• {cleanText(line.replace(/^[-*]\s*/, ''))}</p>;
                                                            }
                                                            // Regular text
                                                            if (line.trim()) {
                                                                return <p key={li} className="text-[var(--text)]">{cleanText(line)}</p>;
                                                            }
                                                            return null;
                                                        })}
                                                    </div>
                                                ) : (
                                                    // Preview
                                                    <p className="line-clamp-2">{result.response?.slice(0, 150)}...</p>
                                                )}
                                            </div>
                                        </div>
                                        {result.mentions?.length > 0 && (
                                            <div className="flex flex-wrap gap-1 mt-2">
                                                {result.mentions.map((m, i) => (
                                                    <span
                                                        key={i}
                                                        className={`text-xs px-2 py-0.5 rounded ${m.entityType === 'brand'
                                                            ? 'bg-green-500/20 text-green-400'
                                                            : 'bg-blue-500/20 text-blue-400'
                                                            }`}
                                                    >
                                                        {m.entityName}
                                                    </span>
                                                ))}
                                            </div>
                                        )}
                                    </>
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Results */}
            {results.length > 0 && (
                <div className="card">
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

                                {/* AI Response with Expand/Collapse */}
                                {result.response && (
                                    <div className="mb-3">
                                        <button
                                            onClick={() => setExpandedResults(prev => ({ ...prev, [`reg-${result.id}`]: !prev[`reg-${result.id}`] }))}
                                            className="text-xs text-[var(--primary)] hover:underline mb-2 flex items-center gap-1"
                                        >
                                            {expandedResults[`reg-${result.id}`] ? '🔽 Collapse Response' : '🔼 Expand Full Response'}
                                        </button>
                                        <div className={`p-3 rounded-lg bg-[var(--surface)] text-sm text-[var(--text)] ${expandedResults[`reg-${result.id}`] ? 'max-h-[500px] overflow-y-auto' : 'max-h-24 overflow-hidden'} transition-all duration-200`}>
                                            {expandedResults[`reg-${result.id}`] ? (
                                                // Full formatted response
                                                <div className="space-y-2">
                                                    {result.response.split('\n').map((line, li) => {
                                                        // Strip all ** and * markers
                                                        const cleanText = (text) => text.replace(/\*\*/g, '').replace(/\*/g, '');

                                                        // Skip separator lines
                                                        if (line.trim() === '---' || line.trim() === '***') {
                                                            return <hr key={li} className="border-[var(--surface-light)] my-3" />;
                                                        }
                                                        // Markdown headings (### Title)
                                                        if (line.startsWith('###')) {
                                                            return <p key={li} className="font-bold text-base text-[var(--primary)] mt-3">{cleanText(line.replace(/^###\s*/, ''))}</p>;
                                                        }
                                                        // Bold headings (lines starting with **)
                                                        if (line.startsWith('**')) {
                                                            return <p key={li} className="font-bold text-[var(--primary)] mt-4">{cleanText(line)}</p>;
                                                        }
                                                        // Regular numbered headings (1. Title:)
                                                        if (/^\d+\.\s/.test(line)) {
                                                            return <p key={li} className="font-semibold text-[var(--primary)] mt-3">{cleanText(line)}</p>;
                                                        }
                                                        // Bullet points (check for - or single * but not **)
                                                        if (line.startsWith('- ') || (line.startsWith('* ') && !line.startsWith('**'))) {
                                                            return <p key={li} className="pl-4 text-[var(--text)]">• {cleanText(line.replace(/^[-*]\s*/, ''))}</p>;
                                                        }
                                                        // Regular text
                                                        if (line.trim()) {
                                                            return <p key={li} className="text-[var(--text)]">{cleanText(line)}</p>;
                                                        }
                                                        return null;
                                                    })}
                                                </div>
                                            ) : (
                                                // Preview
                                                <p className="line-clamp-3">{result.response.substring(0, 200)}...</p>
                                            )}
                                        </div>
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

            {/* Completion Modal */}
            {showCompletionModal && (
                <div className="modal-overlay">
                    <div className="modal-content modal-content-primary text-center">
                        <div className="text-6xl mb-4">🎉</div>
                        <h3 className="text-2xl font-bold text-[var(--text)] mb-2">
                            Analysis Complete!
                        </h3>
                        <p className="text-[var(--text-muted)] mb-6">
                            Your dashboard is ready with fresh AI visibility insights for <strong>{selectedBrand?.name || 'your brand'}</strong>.
                        </p>
                        <div className="flex gap-3 justify-center">
                            <button
                                onClick={() => setShowCompletionModal(false)}
                                className="btn btn-secondary"
                            >
                                Stay Here
                            </button>
                            <button
                                onClick={() => navigate(`/?brand_id=${selectedBrandId}`)}
                                className="btn btn-primary"
                            >
                                View Dashboard →
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Delete Confirmation Modal */}
            {deleteModalPrompt && (
                <div className="modal-overlay">
                    <div className="modal-content modal-content-danger">
                        <div className="flex items-center gap-3 mb-4">
                            <span className="text-3xl">⚠️</span>
                            <h3 className="text-xl font-bold text-[var(--text)]">Delete Question?</h3>
                        </div>
                        <p className="text-[var(--text-muted)] mb-2">Are you sure you want to delete this custom question?</p>
                        <p className="text-sm text-[var(--text)] bg-[var(--background)] p-3 rounded-lg mb-6 font-mono border border-[var(--surface-light)]">
                            "{deleteModalPrompt.template}"
                        </p>
                        <div className="flex gap-3 justify-end">
                            <button
                                onClick={() => setDeleteModalPrompt(null)}
                                className="btn btn-secondary"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={async () => {
                                    try {
                                        await api.deletePrompt(deleteModalPrompt.id);
                                        setTemplates(prev => prev.filter(t => t.id !== deleteModalPrompt.id));
                                        setDeleteModalPrompt(null);
                                    } catch (err) {
                                        console.error('Failed to delete:', err);
                                    }
                                }}
                                className="btn btn-danger"
                            >
                                🗑️ Delete
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
