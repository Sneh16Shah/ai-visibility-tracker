import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'
import * as api from '../api/client'

const demoTrendData = [
    { date: 'Dec 25', visibility: 65, mentions: 12 },
    { date: 'Dec 26', visibility: 72, mentions: 18 },
    { date: 'Dec 27', visibility: 68, mentions: 15 },
    { date: 'Dec 28', visibility: 78, mentions: 22 },
    { date: 'Dec 29', visibility: 82, mentions: 28 },
    { date: 'Dec 30', visibility: 85, mentions: 32 },
    { date: 'Dec 31', visibility: 88, mentions: 35 },
]

// Generate demo chart data based on selected brand
const generateDemoChartData = (brand) => {
    const brandName = brand?.name || 'Your Brand'
    const competitors = brand?.competitors || []
    const colors = ['#10b981', '#f59e0b', '#ef4444', '#8b5cf6']

    const citationData = [
        { name: brandName, value: 35, color: '#6366f1' },
        ...competitors.slice(0, 3).map((c, i) => ({
            name: c.name,
            value: 28 - (i * 6),
            color: colors[i]
        }))
    ]

    const competitorData = [
        { name: brandName, mentions: 35, positive: 28, neutral: 5, negative: 2 },
        ...competitors.slice(0, 3).map((c, i) => ({
            name: c.name,
            mentions: 28 - (i * 6),
            positive: 20 - (i * 5),
            neutral: 6,
            negative: 2
        }))
    ]

    return { citationData, competitorData }
}

// Tooltip style for dark theme
const tooltipStyle = {
    contentStyle: {
        backgroundColor: '#1e293b',
        border: '1px solid #334155',
        borderRadius: '8px',
    },
    itemStyle: { color: '#f1f5f9' },
    labelStyle: { color: '#f1f5f9' }
}

function KPICard({ title, value, subtitle, trend, icon, loading }) {
    return (
        <div className="card card-hover">
            <div className="flex items-start justify-between">
                <div>
                    <p className="text-[var(--text-muted)] text-sm font-medium">{title}</p>
                    {loading ? (
                        <div className="h-9 w-20 bg-[var(--surface-light)] rounded animate-pulse mt-2"></div>
                    ) : (
                        <p className="text-3xl font-bold text-[var(--text)] mt-2">{value}</p>
                    )}
                    {subtitle && (
                        <p className="text-[var(--text-muted)] text-sm mt-1">{subtitle}</p>
                    )}
                </div>
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-[var(--primary)]/20 to-purple-600/20 flex items-center justify-center text-2xl">
                    {icon}
                </div>
            </div>
            {trend !== undefined && trend !== null && (
                <div className={`mt-4 flex items-center gap-1 text-sm ${trend > 0 ? 'text-[var(--success)]' : trend < 0 ? 'text-[var(--error)]' : 'text-[var(--text-muted)]'}`}>
                    <span>{trend > 0 ? '↑' : trend < 0 ? '↓' : '→'}</span>
                    <span>{Math.abs(trend)}% from last week</span>
                </div>
            )}
        </div>
    )
}

export default function Dashboard() {
    const navigate = useNavigate()
    const [searchParams] = useSearchParams()
    const [loading, setLoading] = useState(true)
    const [dashboardData, setDashboardData] = useState(null)
    const [trendData, setTrendData] = useState(demoTrendData)
    const [citationShareData, setCitationShareData] = useState([])
    const [competitorData, setCompetitorData] = useState([])
    const [error, setError] = useState(null)
    const [brands, setBrands] = useState([])
    const [selectedBrandId, setSelectedBrandId] = useState(null)
    const [hasRealData, setHasRealData] = useState(false)
    const [competitorInsights, setCompetitorInsights] = useState({}) // Cache per brand ID
    const [analyzingDeepDive, setAnalyzingDeepDive] = useState(false)
    const [compareModelScores, setCompareModelScores] = useState(null) // Per-model scores from Compare Mode

    // Brand quick actions state
    const [showAddBrand, setShowAddBrand] = useState(false)
    const [newBrandName, setNewBrandName] = useState('')
    const [newBrandIndustry, setNewBrandIndustry] = useState('')
    const [deletingBrandId, setDeletingBrandId] = useState(null)

    // Alert settings state (controlled inputs)
    const [alertThreshold, setAlertThreshold] = useState(0)
    const [scheduleFrequency, setScheduleFrequency] = useState('disabled')

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
                    const selectedBrand = data.brands.find(b => b.id === parseInt(urlBrandId))
                    const demoData = generateDemoChartData(selectedBrand)
                    setCitationShareData(demoData.citationData)
                    setCompetitorData(demoData.competitorData)
                } else if (data.brands && data.brands.length > 0) {
                    setSelectedBrandId(data.brands[0].id)
                    const demoData = generateDemoChartData(data.brands[0])
                    setCitationShareData(demoData.citationData)
                    setCompetitorData(demoData.competitorData)
                }
            } catch (err) {
                console.log('Error fetching brands:', err)
            }
        }
        fetchBrands()
    }, [searchParams])

    // Get selected brand object
    const selectedBrand = brands.find(b => b.id === selectedBrandId)

    // Sync alert settings when brand changes
    useEffect(() => {
        if (selectedBrand) {
            setAlertThreshold(selectedBrand.alert_threshold || 0)
            setScheduleFrequency(selectedBrand.schedule_frequency || 'disabled')
        }
    }, [selectedBrand])

    // Load saved competitor insights when brand changes
    useEffect(() => {
        if (!selectedBrandId) return
        const loadInsights = async () => {
            try {
                const data = await api.getInsights(selectedBrandId)
                if (data.insights) {
                    setCompetitorInsights(prev => ({ ...prev, [selectedBrandId]: data.insights }))
                }
            } catch (err) {
                // No saved insights, that's fine
                console.log('No saved insights for brand:', selectedBrandId)
            }
        }
    }, [selectedBrandId])

    // State for resizable insights panel
    const [insightsExpanded, setInsightsExpanded] = useState(false)

    // Fetch dashboard data when brand changes
    useEffect(() => {
        if (!selectedBrandId) return

        // Reset state when brand changes
        setHasRealData(false)

        const fetchDashboardData = async () => {
            try {
                setLoading(true)
                const data = await api.getDashboardData(selectedBrandId)

                setDashboardData(data)

                // Check if brand has real analysis data (total_mentions > 0)
                const hasData = data.total_mentions && data.total_mentions > 0
                setHasRealData(hasData)

                // Get current brand for fallback data
                const currentBrand = brands.find(b => b.id === selectedBrandId)

                // Update charts - use real data if available, otherwise use brand-specific fallback
                if (hasData) {
                    if (data.trends && data.trends.length > 0) {
                        setTrendData(data.trends.map(t => ({
                            date: new Date(t.snapshot_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
                            visibility: t.visibility_score,
                            mentions: t.mention_count,
                        })))
                    } else {
                        setTrendData(demoTrendData)
                    }

                    // Use API data for charts, or generate brand-specific fallback
                    if (data.citation_breakdown && data.citation_breakdown.length > 0) {
                        setCitationShareData(data.citation_breakdown)
                    } else {
                        const fallback = generateDemoChartData(currentBrand)
                        setCitationShareData(fallback.citationData)
                    }

                    if (data.competitor_data && data.competitor_data.length > 0) {
                        setCompetitorData(data.competitor_data)
                    } else {
                        const fallback = generateDemoChartData(currentBrand)
                        setCompetitorData(fallback.competitorData)
                    }

                    // Set per-model visibility from API
                    if (data.model_visibility && data.model_visibility.length > 0) {
                        setCompareModelScores(data.model_visibility)
                    } else {
                        setCompareModelScores(null)
                    }
                } else {
                    setCompareModelScores(null)
                }

                setError(null)
            } catch (err) {
                console.log('Error fetching dashboard:', err)
                setHasRealData(false)
                setDashboardData(null)
                setError('error')
            } finally {
                setLoading(false)
            }
        }

        fetchDashboardData()

        // Refresh every 30 seconds
        const interval = setInterval(fetchDashboardData, 30000)
        return () => clearInterval(interval)
    }, [selectedBrandId]) // eslint-disable-line react-hooks/exhaustive-deps

    // Handle quick brand creation
    const handleAddBrand = async () => {
        if (!newBrandName.trim()) return
        try {
            const created = await api.createBrand(newBrandName, newBrandIndustry || 'General')
            setBrands(prev => [...prev, created])
            setSelectedBrandId(created.id)
            setNewBrandName('')
            setNewBrandIndustry('')
            setShowAddBrand(false)
        } catch (err) {
            console.error('Failed to create brand:', err)
        }
    }

    // Handle brand deletion
    const handleDeleteBrand = async (brandId) => {
        try {
            await api.deleteBrand(brandId)
            setBrands(prev => prev.filter(b => b.id !== brandId))
            if (selectedBrandId === brandId && brands.length > 1) {
                const remaining = brands.filter(b => b.id !== brandId)
                setSelectedBrandId(remaining[0]?.id || null)
            }
            setDeletingBrandId(null)
        } catch (err) {
            console.error('Failed to delete brand:', err)
        }
    }

    // Empty state when no analysis has been run
    const EmptyState = () => (
        <div className="card text-center p-12">
            <div className="text-6xl mb-4">📊</div>
            <h3 className="text-xl font-semibold text-[var(--text)] mb-2">
                No Analysis Data Yet
            </h3>
            <p className="text-[var(--text-muted)] mb-6 max-w-md mx-auto">
                Run your first AI visibility analysis for <strong>{selectedBrand?.name || 'this brand'}</strong> to see how it appears in AI responses.
            </p>
            <button
                onClick={() => navigate(`/analysis?brand_id=${selectedBrandId}`)}
                className="btn btn-primary btn-primary-lg"
            >
                🚀 Run Analysis
            </button>
        </div>
    )

    return (
        <div className="space-y-5">
            {/* Page Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold text-[var(--text)]">Dashboard</h2>
                    <p className="text-[var(--text-muted)] mt-1">Track your brand's AI visibility metrics</p>
                </div>
                <div className="flex items-center gap-4">
                    {/* Brand Selector */}
                    <select
                        value={selectedBrandId || ''}
                        onChange={(e) => setSelectedBrandId(Number(e.target.value))}
                        className="select"
                    >
                        {brands.length === 0 ? (
                            <option value="">No brands</option>
                        ) : (
                            brands.map(brand => (
                                <option key={brand.id} value={brand.id}>
                                    {brand.name}
                                </option>
                            ))
                        )}
                    </select>

                    {/* Export CSV Button */}
                    {hasRealData && (
                        <button
                            onClick={async () => {
                                try {
                                    await api.exportCSV(selectedBrandId);
                                } catch (err) {
                                    console.error('Export failed:', err);
                                }
                            }}
                            className="btn btn-secondary"
                        >
                            <span>📥</span>
                            <span>Export CSV</span>
                        </button>
                    )}
                </div>
            </div>

            {/* Show empty state or dashboard based on data */}
            {loading ? (
                <div className="flex justify-center items-center h-64">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500"></div>
                </div>
            ) : !hasRealData ? (
                <EmptyState />
            ) : (
                <>
                    {/* KPI Cards - 2 cols mobile, 4 cols desktop */}
                    <div className="grid grid-cols-2 lg:grid-cols-4 gap-2 sm:gap-4">
                        <KPICard
                            title="Visibility Score"
                            value={dashboardData?.visibility_score?.toFixed(0) || '0'}
                            subtitle="out of 100"
                            icon="📊"
                            loading={loading}
                        />
                        <KPICard
                            title="Citation Share"
                            value={`${dashboardData?.citation_share?.toFixed(0) || '0'}%`}
                            subtitle="of AI responses"
                            icon="📈"
                            loading={loading}
                        />
                        <KPICard
                            title="Total Mentions"
                            value={dashboardData?.total_mentions || '0'}
                            subtitle="last 7 days"
                            icon="💬"
                            loading={loading}
                        />
                        <KPICard
                            title="Sentiment Score"
                            value={dashboardData?.sentiment_score?.toFixed(1) || '0'}
                            subtitle="out of 5"
                            icon="😊"
                            loading={loading}
                        />
                    </div>

                    {/* 2-Column Layout: Main Content + Sidebar */}
                    <div className="grid grid-cols-1 lg:grid-cols-[1fr_280px] gap-4 lg:gap-5">

                        {/* Left Column - Main Content */}
                        <div className="space-y-4 lg:space-y-5">
                            {/* Per-Model Visibility Chart - Only shown when Compare Mode has been run */}
                            {compareModelScores && compareModelScores.length > 0 && (
                                <div className="card card-accent-purple">
                                    <div className="flex items-center gap-2 mb-4">
                                        <span className="text-xl">🤖</span>
                                        <h3 className="text-lg font-semibold text-[var(--text)]">Visibility by AI Model</h3>
                                    </div>
                                    <p className="text-sm text-[var(--text-muted)] mb-4">Compare how often your brand is mentioned across different AI assistants</p>
                                    <div className="space-y-3">
                                        {compareModelScores.map((item) => (
                                            <div key={item.modelId} className="flex items-center gap-3">
                                                <div className="w-32 flex items-center gap-2">
                                                    <div
                                                        className="w-3 h-3 rounded-full"
                                                        style={{ backgroundColor: item.color }}
                                                    />
                                                    <span className="text-sm text-[var(--text)]">{item.model}</span>
                                                </div>
                                                <div className="progress-track h-5">
                                                    <div
                                                        className="h-full rounded-full transition-all duration-700"
                                                        style={{
                                                            width: `${item.score}%`,
                                                            backgroundColor: item.color
                                                        }}
                                                    />
                                                </div>
                                                <span className="w-12 text-sm font-mono text-[var(--text)] text-right">{item.score}</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}

                            {/* Competitor Deep Dive */}
                            <div className="card card-accent-orange">
                                <div className="flex items-center justify-between mb-4">
                                    <div className="flex items-center gap-2">
                                        <span className="text-xl">🔍</span>
                                        <h3 className="text-lg font-semibold text-[var(--text)]">Competitor Deep Dive</h3>
                                    </div>
                                    <button
                                        onClick={async () => {
                                            if (!window.puter?.ai) {
                                                alert('Puter.js not loaded');
                                                return;
                                            }
                                            const brand = brands.find(b => b.id === selectedBrandId);
                                            if (!brand?.competitors?.length) {
                                                alert('No competitors configured for this brand');
                                                return;
                                            }
                                            setAnalyzingDeepDive(true);
                                            const competitors = brand.competitors.map(c => c.name).join(', ');
                                            try {
                                                const insights = await window.puter.ai.chat(
                                                    `Analyze why competitors (${competitors}) might rank better than "${brand.name}" in AI assistant responses. 

Format your response with these sections:
## Why Competitors Rank Higher
- List 3-4 key reasons with specific examples

## Actionable Recommendations for ${brand.name}
- List 5 specific, actionable steps to improve AI visibility
- Include SEO, content strategy, structured data, and brand authority tips

Keep each point concise (1-2 sentences). Industry: ${brand.industry || 'Technology'}`,
                                                    { model: 'gpt-5.2' }
                                                );
                                                const result = typeof insights === 'string' ? insights : insights.message?.content || JSON.stringify(insights);
                                                setCompetitorInsights(prev => ({ ...prev, [selectedBrandId]: result }));

                                                // Save to database
                                                try {
                                                    await api.saveInsights(selectedBrandId, result);
                                                } catch (saveErr) {
                                                    console.log('Could not save insights to DB:', saveErr);
                                                }
                                            } catch (err) {
                                                console.error('Deep dive failed:', err);
                                            } finally {
                                                setAnalyzingDeepDive(false);
                                            }
                                        }}
                                        disabled={analyzingDeepDive}
                                        className="px-3 py-1.5 text-sm badge badge-amber cursor-pointer hover:opacity-80 transition-opacity flex items-center gap-1 disabled:opacity-50"
                                    >
                                        {analyzingDeepDive ? (
                                            <><span className="spinner">⏳</span> Analyzing...</>
                                        ) : (
                                            <><span>🤖</span> Analyze</>
                                        )}
                                    </button>
                                </div>
                                <p className="text-sm text-[var(--text-muted)] mb-4">Get AI-powered insights on competitor advantages and actionable tips to improve your visibility</p>

                                {analyzingDeepDive ? (
                                    <div className="p-6 rounded-xl bg-[var(--background)] border border-[var(--surface-light)] text-center">
                                        <div className="animate-spin text-3xl mb-3">🔄</div>
                                        <p className="text-[var(--text-muted)]">Analyzing competitors and generating recommendations...</p>
                                    </div>
                                ) : competitorInsights[selectedBrandId] ? (
                                    <div className="relative">
                                        {/* Expand/Collapse Toggle */}
                                        <button
                                            onClick={() => setInsightsExpanded(!insightsExpanded)}
                                            className="absolute top-2 right-2 z-10 p-1.5 rounded-lg bg-[var(--surface)] border border-[var(--surface-light)] hover:bg-[var(--surface-light)] transition-colors text-xs"
                                            title={insightsExpanded ? 'Collapse' : 'Expand'}
                                        >
                                            {insightsExpanded ? '🔽 Collapse' : '🔼 Expand'}
                                        </button>
                                        <div className={`p-4 rounded-xl bg-[var(--background)] border border-[var(--surface-light)] text-sm text-[var(--text)] space-y-3 overflow-y-auto transition-all ${insightsExpanded ? 'max-h-[600px]' : 'max-h-52'}`}>
                                            {competitorInsights[selectedBrandId].split('\n').map((line, i) => {
                                                if (line.startsWith('## ')) {
                                                    return <h4 key={i} className="font-bold text-base text-[var(--primary)] mt-3 first:mt-0">{line.replace('## ', '')}</h4>;
                                                } else if (line.startsWith('- **') || line.startsWith('* **')) {
                                                    const parts = line.replace(/^[-*]\s*/, '').split('**');
                                                    return (
                                                        <div key={i} className="flex gap-2 items-start pl-2">
                                                            <span className="text-[var(--primary)]">•</span>
                                                            <span><strong>{parts[1]}</strong>{parts[2] || ''}</span>
                                                        </div>
                                                    );
                                                } else if (line.startsWith('- ') || line.startsWith('* ')) {
                                                    return (
                                                        <div key={i} className="flex gap-2 items-start pl-2">
                                                            <span className="text-[var(--primary)]">•</span>
                                                            <span>{line.replace(/^[-*]\s*/, '')}</span>
                                                        </div>
                                                    );
                                                } else if (line.trim()) {
                                                    return <p key={i} className="text-[var(--text-muted)]">{line}</p>;
                                                }
                                                return null;
                                            })}
                                        </div>
                                    </div>
                                ) : (
                                    <div className="p-4 rounded-xl bg-[var(--background)] border border-dashed border-[var(--surface-light)] text-center text-[var(--text-muted)]">
                                        Click "Analyze" to get AI-powered insights and actionable recommendations
                                    </div>
                                )}
                            </div>

                            {/* Charts Row - Side by Side */}
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                {/* Visibility Trend */}
                                <div className="card">
                                    <h3 className="text-base font-semibold text-[var(--text)] mb-3">Visibility Trend</h3>
                                    <ResponsiveContainer width="100%" height={200}>
                                        <LineChart data={trendData}>
                                            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                                            <XAxis dataKey="date" stroke="#94a3b8" fontSize={12} />
                                            <YAxis stroke="#94a3b8" fontSize={12} />
                                            <Tooltip {...tooltipStyle} />
                                            <Line
                                                type="monotone"
                                                dataKey="visibility"
                                                stroke="var(--primary)"
                                                strokeWidth={3}
                                                dot={{ fill: 'var(--primary)', strokeWidth: 2, r: 4 }}
                                                activeDot={{ r: 6, fill: 'var(--primary-light)' }}
                                            />
                                        </LineChart>
                                    </ResponsiveContainer>
                                </div>

                                {/* Citation Share Pie */}
                                <div className="card">
                                    <h3 className="text-base font-semibold text-[var(--text)] mb-3">Citation Share</h3>
                                    <ResponsiveContainer width="100%" height={200}>
                                        <PieChart>
                                            <Pie
                                                data={citationShareData}
                                                cx="50%"
                                                cy="50%"
                                                innerRadius={50}
                                                outerRadius={80}
                                                paddingAngle={5}
                                                dataKey="value"
                                            >
                                                {citationShareData.map((entry, index) => (
                                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                                ))}
                                            </Pie>
                                            <Tooltip {...tooltipStyle} />
                                        </PieChart>
                                    </ResponsiveContainer>
                                    {/* Legend */}
                                    <div className="flex flex-wrap justify-center gap-3 mt-2">
                                        {citationShareData.map((item) => (
                                            <div key={item.name} className="flex items-center gap-2">
                                                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                                                <span className="text-sm text-[var(--text-muted)]">{item.name}</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            </div>

                            {/* Competitor Comparison */}
                            <div className="card">
                                <h3 className="text-base font-semibold text-[var(--text)] mb-3">Brand vs Competitors</h3>
                                <ResponsiveContainer width="100%" height={180}>
                                    <BarChart data={competitorData} layout="vertical">
                                        <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                                        <XAxis type="number" stroke="#94a3b8" fontSize={12} />
                                        <YAxis dataKey="name" type="category" stroke="#94a3b8" fontSize={12} width={100} />
                                        <Tooltip {...tooltipStyle} />
                                        <Bar dataKey="positive" stackId="a" fill="var(--success)" name="Positive" radius={[0, 0, 0, 0]} />
                                        <Bar dataKey="neutral" stackId="a" fill="var(--warning)" name="Neutral" />
                                        <Bar dataKey="negative" stackId="a" fill="var(--error)" name="Negative" radius={[0, 4, 4, 0]} />
                                    </BarChart>
                                </ResponsiveContainer>
                            </div>
                        </div>
                        {/* End Left Column */}

                        {/* Right Sidebar - Hidden on mobile for cleaner experience */}
                        <div className="hidden lg:block space-y-4">
                            {/* Brand Quick Actions */}
                            <div className="card">
                                <div className="flex items-center gap-2 mb-4">
                                    <span className="text-xl">🏢</span>
                                    <h3 className="text-lg font-semibold text-[var(--text)]">Brand Actions</h3>
                                </div>

                                {/* Current Brand */}
                                <div className="mb-4 p-3 rounded-lg bg-[var(--background)] border border-[var(--primary)]/30">
                                    <p className="text-xs text-[var(--text-muted)] mb-1">Current Brand</p>
                                    <p className="text-[var(--text)] font-medium">{selectedBrand?.name || 'None'}</p>
                                    <p className="text-xs text-[var(--text-muted)]">{selectedBrand?.industry || ''}</p>
                                </div>

                                {/* Brand List */}
                                <div className="space-y-2 mb-4 max-h-40 overflow-y-auto">
                                    {brands.map(brand => (
                                        <div
                                            key={brand.id}
                                            onClick={() => setSelectedBrandId(brand.id)}
                                            className={`flex items-center justify-between p-2 rounded-lg cursor-pointer transition-colors ${brand.id === selectedBrandId
                                                ? 'bg-[var(--primary)]/20 border border-[var(--primary)]/30'
                                                : 'hover:bg-[var(--surface-light)]'
                                                }`}
                                        >
                                            <span className="text-sm text-[var(--text)]">{brand.name}</span>
                                            {brand.id !== selectedBrandId && (
                                                <button
                                                    onClick={(e) => {
                                                        e.stopPropagation()
                                                        setDeletingBrandId(brand.id)
                                                    }}
                                                    className="text-xs text-[var(--text-muted)] hover:text-[var(--error)]"
                                                >
                                                    🗑️
                                                </button>
                                            )}
                                        </div>
                                    ))}
                                </div>

                                {/* Add Brand Form */}
                                {showAddBrand ? (
                                    <div className="space-y-2">
                                        <input
                                            type="text"
                                            placeholder="Brand name"
                                            value={newBrandName}
                                            onChange={(e) => setNewBrandName(e.target.value)}
                                            className="input"
                                        />
                                        <input
                                            type="text"
                                            placeholder="Industry (optional)"
                                            value={newBrandIndustry}
                                            onChange={(e) => setNewBrandIndustry(e.target.value)}
                                            className="input"
                                        />
                                        <div className="flex gap-2">
                                            <button onClick={handleAddBrand} className="btn btn-primary flex-1 text-sm">Add</button>
                                            <button onClick={() => setShowAddBrand(false)} className="btn btn-secondary text-sm">Cancel</button>
                                        </div>
                                    </div>
                                ) : (
                                    <button
                                        onClick={() => setShowAddBrand(true)}
                                        className="w-full py-2 border-2 border-dashed border-[var(--surface-light)] rounded-lg text-[var(--text-muted)] hover:border-[var(--primary)] hover:text-[var(--primary)] transition-colors text-sm"
                                    >
                                        + Add New Brand
                                    </button>
                                )}
                            </div>

                            {/* Alert & Schedule Settings - Compact */}
                            <div className="card card-accent-cyan">
                                <div className="flex items-center gap-2 mb-4">
                                    <span className="text-xl">⚙️</span>
                                    <h3 className="text-lg font-semibold text-[var(--text)]">Settings</h3>
                                </div>

                                <div className="space-y-4">
                                    <div>
                                        <label className="label">Alert Threshold</label>
                                        <div className="flex items-center gap-2">
                                            <input
                                                type="number"
                                                min="0"
                                                max="100"
                                                value={alertThreshold}
                                                onChange={(e) => setAlertThreshold(Number(e.target.value))}
                                                className="input w-20"
                                            />
                                            <span className="text-[var(--text-muted)] text-sm">/ 100</span>
                                        </div>
                                    </div>

                                    <div>
                                        <label className="label">Auto-Run</label>
                                        <select
                                            value={scheduleFrequency}
                                            onChange={(e) => setScheduleFrequency(e.target.value)}
                                            className="select w-full"
                                        >
                                            <option value="disabled">Disabled</option>
                                            <option value="daily">Daily</option>
                                            <option value="weekly">Weekly</option>
                                        </select>
                                    </div>

                                    <button
                                        onClick={async () => {
                                            try {
                                                await api.updateAlertSettings(selectedBrandId, alertThreshold, scheduleFrequency)
                                                alert('Settings saved!')
                                            } catch (err) {
                                                console.error('Failed to save:', err)
                                            }
                                        }}
                                        className="btn w-full bg-[var(--info)]/20 text-[var(--info)] hover:bg-[var(--info)]/30"
                                    >
                                        💾 Save
                                    </button>
                                </div>
                            </div>

                            {/* Quick Actions */}
                            <div className="card">
                                <button
                                    onClick={() => navigate(`/analysis?brand_id=${selectedBrandId}`)}
                                    className="btn btn-primary w-full mb-2"
                                >
                                    🚀 Run Analysis
                                </button>
                                <button
                                    onClick={() => navigate('/brands')}
                                    className="btn btn-secondary w-full text-sm"
                                >
                                    Manage Brands →
                                </button>
                            </div>
                        </div>
                        {/* End Right Sidebar */}
                    </div>
                    {/* End 2-Column Layout */}
                </>
            )}

            {/* Delete Brand Confirmation Modal */}
            {deletingBrandId && (
                <div className="modal-overlay">
                    <div className="modal-content modal-content-danger">
                        <h3 className="text-lg font-bold text-[var(--text)] mb-2">Delete Brand?</h3>
                        <p className="text-[var(--text-muted)] mb-4">
                            This will delete "{brands.find(b => b.id === deletingBrandId)?.name}" and all its analysis data.
                        </p>
                        <div className="flex gap-2 justify-end">
                            <button onClick={() => setDeletingBrandId(null)} className="btn btn-secondary">Cancel</button>
                            <button onClick={() => handleDeleteBrand(deletingBrandId)} className="btn btn-danger">Delete</button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
