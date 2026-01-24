import { useState, useEffect } from 'react'
import * as api from '../api/client'

export default function BrandSetup() {
    const [brands, setBrands] = useState([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState(null)
    const [newBrand, setNewBrand] = useState({
        name: '',
        industry: '',
        aliases: '',
        competitors: '',
    })

    // Edit and delete modal state
    const [editingBrand, setEditingBrand] = useState(null)
    const [editForm, setEditForm] = useState({ name: '', industry: '' })
    const [deleteModalBrand, setDeleteModalBrand] = useState(null)

    // Fetch brands on mount
    useEffect(() => {
        fetchBrands()
    }, [])

    const fetchBrands = async () => {
        try {
            setLoading(true)
            const data = await api.getBrands()
            setBrands(data.brands || [])
            setError(null)
        } catch (err) {
            console.log('Error fetching brands:', err)
            // Use demo data if API fails
            setBrands([{
                id: 1,
                name: 'Acme Corp',
                industry: 'Technology',
                aliases: [{ id: 1, alias: 'Acme' }, { id: 2, alias: 'ACME Corporation' }],
                competitors: [{ id: 1, name: 'TechGiant' }, { id: 2, name: 'InnovateCo' }],
            }])
            setError('Running in demo mode')
        } finally {
            setLoading(false)
        }
    }

    const handleSubmit = async (e) => {
        e.preventDefault()
        setSaving(true)
        setError(null)

        try {
            const brandData = {
                name: newBrand.name,
                industry: newBrand.industry,
                aliases: newBrand.aliases.split(',').map(a => a.trim()).filter(Boolean),
                competitors: newBrand.competitors.split(',').map(c => c.trim()).filter(Boolean),
            }

            await api.createBrand(brandData)
            await fetchBrands()
            setNewBrand({ name: '', industry: '', aliases: '', competitors: '' })
            setShowForm(false)
        } catch (err) {
            console.error('Error creating brand:', err)
            setError(err.message || 'Failed to create brand')

            // Fallback: add to local state for demo
            const brand = {
                id: brands.length + 1,
                name: newBrand.name,
                industry: newBrand.industry,
                aliases: newBrand.aliases.split(',').map((a, i) => ({ id: i + 1, alias: a.trim() })).filter(a => a.alias),
                competitors: newBrand.competitors.split(',').map((c, i) => ({ id: i + 1, name: c.trim() })).filter(c => c.name),
            }
            setBrands([...brands, brand])
            setNewBrand({ name: '', industry: '', aliases: '', competitors: '' })
            setShowForm(false)
        } finally {
            setSaving(false)
        }
    }

    const deleteBrand = async (id) => {
        try {
            await api.deleteBrand(id)
            setBrands(brands.filter(b => b.id !== id))
        } catch (err) {
            console.error('Error deleting brand:', err)
            // Fallback: remove from local state
            setBrands(brands.filter(b => b.id !== id))
        }
    }

    if (loading) {
        return (
            <div className="space-y-8">
                <div className="h-10 w-48 bg-[var(--surface)] rounded animate-pulse"></div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    {[1, 2].map(i => (
                        <div key={i} className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)] h-48 animate-pulse"></div>
                    ))}
                </div>
            </div>
        )
    }

    return (
        <div className="space-y-8">
            {/* Page Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold text-[var(--text)]">Brand Setup</h2>
                    <p className="text-[var(--text-muted)] mt-1">Configure brands and competitors to track</p>
                </div>
                <div className="flex items-center gap-4">
                    {error && (
                        <span className="text-amber-400 text-sm">{error}</span>
                    )}
                    <button
                        onClick={() => setShowForm(!showForm)}
                        className="px-4 py-2 bg-[var(--primary)] hover:bg-[var(--primary-dark)] text-white rounded-lg font-medium transition-colors duration-200 flex items-center gap-2"
                    >
                        <span>+</span>
                        <span>Add Brand</span>
                    </button>
                </div>
            </div>

            {/* Add Brand Form */}
            {showForm && (
                <div className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)]">
                    <h3 className="text-lg font-semibold text-[var(--text)] mb-4">Add New Brand</h3>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-[var(--text-muted)] mb-2">
                                    Brand Name *
                                </label>
                                <input
                                    type="text"
                                    required
                                    value={newBrand.name}
                                    onChange={(e) => setNewBrand({ ...newBrand, name: e.target.value })}
                                    className="w-full px-4 py-2 bg-[var(--background)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)] transition-colors"
                                    placeholder="e.g., Acme Corp"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-[var(--text-muted)] mb-2">
                                    Industry / Category *
                                </label>
                                <input
                                    type="text"
                                    required
                                    value={newBrand.industry}
                                    onChange={(e) => setNewBrand({ ...newBrand, industry: e.target.value })}
                                    className="w-full px-4 py-2 bg-[var(--background)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)] transition-colors"
                                    placeholder="e.g., Technology, SaaS"
                                />
                            </div>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-[var(--text-muted)] mb-2">
                                Brand Aliases (comma-separated)
                            </label>
                            <input
                                type="text"
                                value={newBrand.aliases}
                                onChange={(e) => setNewBrand({ ...newBrand, aliases: e.target.value })}
                                className="w-full px-4 py-2 bg-[var(--background)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)] transition-colors"
                                placeholder="e.g., Acme, ACME Inc"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-[var(--text-muted)] mb-2">
                                Competitors (comma-separated)
                            </label>
                            <input
                                type="text"
                                value={newBrand.competitors}
                                onChange={(e) => setNewBrand({ ...newBrand, competitors: e.target.value })}
                                className="w-full px-4 py-2 bg-[var(--background)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)] transition-colors"
                                placeholder="e.g., CompetitorA, CompetitorB"
                            />
                        </div>
                        <div className="flex gap-3">
                            <button
                                type="submit"
                                disabled={saving}
                                className="px-6 py-2 bg-[var(--primary)] hover:bg-[var(--primary-dark)] text-white rounded-lg font-medium transition-colors duration-200 disabled:opacity-50"
                            >
                                {saving ? 'Saving...' : 'Save Brand'}
                            </button>
                            <button
                                type="button"
                                onClick={() => setShowForm(false)}
                                className="px-6 py-2 bg-[var(--surface-light)] hover:bg-[var(--surface)] text-[var(--text)] rounded-lg font-medium transition-colors duration-200"
                            >
                                Cancel
                            </button>
                        </div>
                    </form>
                </div>
            )}

            {/* Brand Cards */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {brands.map((brand) => (
                    <div
                        key={brand.id}
                        className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--surface-light)] hover:border-[var(--primary)] transition-all duration-300"
                    >
                        <div className="flex items-start justify-between mb-4">
                            <div>
                                <h3 className="text-xl font-bold text-[var(--text)]">{brand.name}</h3>
                                <span className="inline-block px-3 py-1 bg-[var(--primary)]/20 text-[var(--primary)] rounded-full text-sm mt-2">
                                    {brand.industry}
                                </span>
                            </div>
                            <div className="flex gap-1">
                                <button
                                    onClick={() => {
                                        setEditingBrand(brand)
                                        setEditForm({ name: brand.name, industry: brand.industry })
                                    }}
                                    className="text-blue-400 hover:text-blue-300 transition-colors p-2"
                                    title="Edit"
                                >
                                    ✏️
                                </button>
                                <button
                                    onClick={() => setDeleteModalBrand(brand)}
                                    className="text-red-400 hover:text-red-300 transition-colors p-2"
                                    title="Delete"
                                >
                                    🗑️
                                </button>
                            </div>
                        </div>

                        {/* Aliases */}
                        <div className="mb-4">
                            <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Aliases</p>
                            <div className="flex flex-wrap gap-2">
                                {(brand.aliases || []).map((alias, i) => (
                                    <span
                                        key={alias.id || i}
                                        className="px-3 py-1 bg-[var(--surface-light)] text-[var(--text)] rounded-lg text-sm"
                                    >
                                        {alias.alias || alias}
                                    </span>
                                ))}
                                {(!brand.aliases || brand.aliases.length === 0) && (
                                    <span className="text-[var(--text-muted)] text-sm">No aliases</span>
                                )}
                            </div>
                        </div>

                        {/* Competitors */}
                        <div>
                            <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Competitors</p>
                            <div className="flex flex-wrap gap-2">
                                {(brand.competitors || []).map((comp, i) => (
                                    <span
                                        key={comp.id || i}
                                        className="px-3 py-1 bg-amber-500/20 text-amber-400 rounded-lg text-sm"
                                    >
                                        {comp.name || comp}
                                    </span>
                                ))}
                                {(!brand.competitors || brand.competitors.length === 0) && (
                                    <span className="text-[var(--text-muted)] text-sm">No competitors</span>
                                )}
                            </div>
                        </div>
                    </div>
                ))}
            </div>

            {/* Empty State */}
            {brands.length === 0 && (
                <div className="text-center py-16">
                    <div className="text-6xl mb-4">🏢</div>
                    <h3 className="text-xl font-semibold text-[var(--text)] mb-2">No brands configured</h3>
                    <p className="text-[var(--text-muted)] mb-6">Add your first brand to start tracking AI visibility</p>
                    <button
                        onClick={() => setShowForm(true)}
                        className="px-6 py-3 bg-[var(--primary)] hover:bg-[var(--primary-dark)] text-white rounded-lg font-medium transition-colors duration-200"
                    >
                        Add Your First Brand
                    </button>
                </div>
            )}

            {/* Edit Brand Modal */}
            {editingBrand && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 backdrop-blur-sm">
                    <div className="bg-[var(--surface)] rounded-2xl p-6 border border-[var(--primary)]/30 max-w-md w-full mx-4">
                        <div className="flex items-center gap-3 mb-4">
                            <span className="text-2xl">✏️</span>
                            <h3 className="text-xl font-bold text-[var(--text)]">Edit Brand</h3>
                        </div>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-[var(--text-muted)] mb-2">Brand Name</label>
                                <input
                                    type="text"
                                    value={editForm.name}
                                    onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                                    className="w-full px-4 py-2 bg-[var(--background)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)]"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-[var(--text-muted)] mb-2">Industry</label>
                                <input
                                    type="text"
                                    value={editForm.industry}
                                    onChange={(e) => setEditForm({ ...editForm, industry: e.target.value })}
                                    className="w-full px-4 py-2 bg-[var(--background)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)]"
                                />
                            </div>
                        </div>
                        <div className="flex gap-3 justify-end mt-6">
                            <button
                                onClick={() => setEditingBrand(null)}
                                className="px-4 py-2 border border-[var(--surface-light)] text-[var(--text-muted)] rounded-lg hover:bg-[var(--surface-light)] transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={async () => {
                                    try {
                                        await api.updateBrand(editingBrand.id, editForm)
                                        setBrands(brands.map(b =>
                                            b.id === editingBrand.id ? { ...b, ...editForm } : b
                                        ))
                                        setEditingBrand(null)
                                    } catch (err) {
                                        console.error('Failed to update brand:', err)
                                        // Fallback: update local state
                                        setBrands(brands.map(b =>
                                            b.id === editingBrand.id ? { ...b, ...editForm } : b
                                        ))
                                        setEditingBrand(null)
                                    }
                                }}
                                className="px-4 py-2 bg-[var(--primary)] text-white font-medium rounded-lg hover:bg-[var(--primary-dark)] transition-colors"
                            >
                                Save Changes
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Delete Confirmation Modal */}
            {deleteModalBrand && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 backdrop-blur-sm">
                    <div className="bg-[var(--surface)] rounded-2xl p-6 border border-red-500/30 max-w-md mx-4">
                        <div className="flex items-center gap-3 mb-4">
                            <span className="text-3xl">⚠️</span>
                            <h3 className="text-xl font-bold text-[var(--text)]">Delete Brand?</h3>
                        </div>
                        <p className="text-[var(--text-muted)] mb-2">Are you sure you want to delete this brand?</p>
                        <p className="text-lg text-[var(--text)] font-semibold mb-2">"{deleteModalBrand.name}"</p>
                        <p className="text-sm text-red-400 mb-6">⚠️ This will also delete all analysis results for this brand.</p>
                        <div className="flex gap-3 justify-end">
                            <button
                                onClick={() => setDeleteModalBrand(null)}
                                className="px-4 py-2 border border-[var(--surface-light)] text-[var(--text-muted)] rounded-lg hover:bg-[var(--surface-light)] transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={async () => {
                                    await deleteBrand(deleteModalBrand.id)
                                    setDeleteModalBrand(null)
                                }}
                                className="px-4 py-2 bg-red-500 text-white font-medium rounded-lg hover:bg-red-600 transition-colors"
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
