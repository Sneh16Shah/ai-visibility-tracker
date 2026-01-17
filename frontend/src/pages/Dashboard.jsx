import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
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
        <div className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)] hover:border-[var(--primary)] transition-all duration-300">
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
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-indigo-500/20 to-purple-600/20 flex items-center justify-center text-2xl">
                    {icon}
                </div>
            </div>
            {trend !== undefined && trend !== null && (
                <div className={`mt-4 flex items-center gap-1 text-sm ${trend > 0 ? 'text-green-400' : trend < 0 ? 'text-red-400' : 'text-[var(--text-muted)]'}`}>
                    <span>{trend > 0 ? '↑' : trend < 0 ? '↓' : '→'}</span>
                    <span>{Math.abs(trend)}% from last week</span>
                </div>
            )}
        </div>
    )
}

export default function Dashboard() {
    const navigate = useNavigate()
    const [loading, setLoading] = useState(true)
    const [dashboardData, setDashboardData] = useState(null)
    const [trendData, setTrendData] = useState(demoTrendData)
    const [citationShareData, setCitationShareData] = useState([])
    const [competitorData, setCompetitorData] = useState([])
    const [error, setError] = useState(null)
    const [brands, setBrands] = useState([])
    const [selectedBrandId, setSelectedBrandId] = useState(null)
    const [hasRealData, setHasRealData] = useState(false)

    // Fetch brands on mount
    useEffect(() => {
        const fetchBrands = async () => {
            try {
                const data = await api.getBrands()
                setBrands(data.brands || [])
                if (data.brands && data.brands.length > 0) {
                    setSelectedBrandId(data.brands[0].id)
                    // Set initial demo data with real brand names
                    const demoData = generateDemoChartData(data.brands[0])
                    setCitationShareData(demoData.citationData)
                    setCompetitorData(demoData.competitorData)
                }
            } catch (err) {
                console.log('Error fetching brands:', err)
            }
        }
        fetchBrands()
    }, [])

    // Get selected brand object
    const selectedBrand = brands.find(b => b.id === selectedBrandId)

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
    }, [selectedBrandId, brands])

    // Empty state when no analysis has been run
    const EmptyState = () => (
        <div className="bg-[var(--surface)] rounded-2xl p-12 border border-[var(--surface-light)] text-center">
            <div className="text-6xl mb-4">📊</div>
            <h3 className="text-xl font-semibold text-[var(--text)] mb-2">
                No Analysis Data Yet
            </h3>
            <p className="text-[var(--text-muted)] mb-6 max-w-md mx-auto">
                Run your first AI visibility analysis for <strong>{selectedBrand?.name || 'this brand'}</strong> to see how it appears in AI responses.
            </p>
            <button
                onClick={() => navigate('/analysis')}
                className="px-6 py-3 bg-gradient-to-r from-indigo-500 to-purple-600 text-white font-medium rounded-xl hover:opacity-90 transition-all duration-300"
            >
                🚀 Run Analysis
            </button>
        </div>
    )

    return (
        <div className="space-y-8">
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
                        className="px-4 py-2 bg-[var(--surface)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)]"
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
                    {/* KPI Cards */}
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
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


                    {/* Charts Row */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                        {/* Visibility Trend */}
                        <div className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)]">
                            <h3 className="text-lg font-semibold text-[var(--text)] mb-4">Visibility Trend</h3>
                            <ResponsiveContainer width="100%" height={300}>
                                <LineChart data={trendData}>
                                    <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                                    <XAxis dataKey="date" stroke="#94a3b8" fontSize={12} />
                                    <YAxis stroke="#94a3b8" fontSize={12} />
                                    <Tooltip {...tooltipStyle} />
                                    <Line
                                        type="monotone"
                                        dataKey="visibility"
                                        stroke="#6366f1"
                                        strokeWidth={3}
                                        dot={{ fill: '#6366f1', strokeWidth: 2, r: 4 }}
                                        activeDot={{ r: 6, fill: '#818cf8' }}
                                    />
                                </LineChart>
                            </ResponsiveContainer>
                        </div>

                        {/* Citation Share Pie */}
                        <div className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)]">
                            <h3 className="text-lg font-semibold text-[var(--text)] mb-4">Citation Share</h3>
                            <ResponsiveContainer width="100%" height={300}>
                                <PieChart>
                                    <Pie
                                        data={citationShareData}
                                        cx="50%"
                                        cy="50%"
                                        innerRadius={60}
                                        outerRadius={100}
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
                            <div className="flex flex-wrap justify-center gap-4 mt-4">
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
                    <div className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)]">
                        <h3 className="text-lg font-semibold text-[var(--text)] mb-4">Brand vs Competitors</h3>
                        <ResponsiveContainer width="100%" height={300}>
                            <BarChart data={competitorData} layout="vertical">
                                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                                <XAxis type="number" stroke="#94a3b8" fontSize={12} />
                                <YAxis dataKey="name" type="category" stroke="#94a3b8" fontSize={12} width={100} />
                                <Tooltip {...tooltipStyle} />
                                <Bar dataKey="positive" stackId="a" fill="#10b981" name="Positive" radius={[0, 0, 0, 0]} />
                                <Bar dataKey="neutral" stackId="a" fill="#f59e0b" name="Neutral" />
                                <Bar dataKey="negative" stackId="a" fill="#ef4444" name="Negative" radius={[0, 4, 4, 0]} />
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                </>
            )}
        </div>
    )
}
