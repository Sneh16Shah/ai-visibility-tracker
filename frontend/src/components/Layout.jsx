import { NavLink, useNavigate } from 'react-router-dom'
import { useState, useEffect } from 'react'
import * as api from '../api/client'

export default function Layout({ children }) {
    const navigate = useNavigate()
    const [user, setUser] = useState(null)

    useEffect(() => {
        // Check for logged in user
        const currentUser = api.getCurrentUser()
        setUser(currentUser)
    }, [])

    const handleLogout = () => {
        api.logout()
        setUser(null)
        navigate('/login')
    }

    return (
        <div className="min-h-screen bg-[var(--background)]">
            {/* Header */}
            <header className="bg-[var(--surface)] border-b border-[var(--surface-light)] sticky top-0 z-50">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex items-center justify-between h-16">
                        {/* Logo */}
                        <div className="flex items-center gap-3">
                            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center">
                                <span className="text-white font-bold text-lg">AI</span>
                            </div>
                            <h1 className="text-xl font-bold text-[var(--text)]">
                                Visibility Tracker
                            </h1>
                        </div>

                        {/* Navigation */}
                        <nav className="flex items-center gap-1">
                            <NavLink
                                to="/"
                                className={({ isActive }) =>
                                    `px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${isActive
                                        ? 'bg-[var(--primary)] text-white'
                                        : 'text-[var(--text-muted)] hover:text-[var(--text)] hover:bg-[var(--surface-light)]'
                                    }`
                                }
                            >
                                Dashboard
                            </NavLink>
                            <NavLink
                                to="/brands"
                                className={({ isActive }) =>
                                    `px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${isActive
                                        ? 'bg-[var(--primary)] text-white'
                                        : 'text-[var(--text-muted)] hover:text-[var(--text)] hover:bg-[var(--surface-light)]'
                                    }`
                                }
                            >
                                Brands
                            </NavLink>
                            <NavLink
                                to="/analysis"
                                className={({ isActive }) =>
                                    `px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${isActive
                                        ? 'bg-[var(--primary)] text-white'
                                        : 'text-[var(--text-muted)] hover:text-[var(--text)] hover:bg-[var(--surface-light)]'
                                    }`
                                }
                            >
                                Run Analysis
                            </NavLink>
                        </nav>

                        {/* User Menu */}
                        <div className="flex items-center gap-3">
                            {user ? (
                                <>
                                    <span className="text-sm text-[var(--text-muted)]">
                                        {user.name || user.email}
                                    </span>
                                    <button
                                        onClick={handleLogout}
                                        className="px-3 py-1.5 text-sm text-[var(--text-muted)] hover:text-[var(--text)] hover:bg-[var(--surface-light)] rounded-lg transition-colors"
                                    >
                                        Logout
                                    </button>
                                </>
                            ) : (
                                <NavLink
                                    to="/login"
                                    className="px-4 py-2 bg-[var(--primary)] text-white rounded-lg text-sm font-medium hover:bg-[var(--primary-dark)] transition-colors"
                                >
                                    Sign In
                                </NavLink>
                            )}
                        </div>
                    </div>
                </div>
            </header>

            {/* Main Content */}
            <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {children}
            </main>
        </div>
    )
}
