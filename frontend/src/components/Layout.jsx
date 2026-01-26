import { NavLink, useNavigate } from 'react-router-dom'
import { useState, useEffect } from 'react'
import * as api from '../api/client'
import { useTheme } from '../contexts/ThemeContext'

export default function Layout({ children }) {
    const navigate = useNavigate()
    const [user, setUser] = useState(null)
    const { theme, toggleTheme, isDark } = useTheme()

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
                <div className="max-w-screen-2xl mx-auto px-6 lg:px-10">
                    <div className="flex items-center justify-between h-16">
                        {/* Logo */}
                        <div className="flex items-center gap-3">
                            <div className="logo-icon">AI</div>
                            <h1 className="text-xl font-bold text-[var(--text)]">
                                Visibility Tracker
                            </h1>
                        </div>

                        {/* Navigation */}
                        <nav className="flex items-center gap-1">
                            <NavLink
                                to="/"
                                className={({ isActive }) =>
                                    `nav-link ${isActive ? 'nav-link-active' : ''}`
                                }
                            >
                                Dashboard
                            </NavLink>
                            <NavLink
                                to="/brands"
                                className={({ isActive }) =>
                                    `nav-link ${isActive ? 'nav-link-active' : ''}`
                                }
                            >
                                Brands
                            </NavLink>
                            <NavLink
                                to="/analysis"
                                className={({ isActive }) =>
                                    `nav-link ${isActive ? 'nav-link-active' : ''}`
                                }
                            >
                                Run Analysis
                            </NavLink>
                        </nav>

                        {/* User Menu */}
                        <div className="flex items-center gap-3">
                            {/* Theme Toggle */}
                            <button
                                onClick={toggleTheme}
                                className="theme-toggle"
                                title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
                            >
                                {isDark ? '☀️' : '🌙'}
                            </button>

                            {user ? (
                                <>
                                    <span className="text-sm text-[var(--text-muted)]">
                                        {user.name || user.email}
                                    </span>
                                    <button
                                        onClick={handleLogout}
                                        className="btn btn-ghost text-sm"
                                    >
                                        Logout
                                    </button>
                                </>
                            ) : (
                                <NavLink
                                    to="/login"
                                    className="btn btn-primary text-sm"
                                >
                                    Sign In
                                </NavLink>
                            )}
                        </div>
                    </div>
                </div>
            </header>

            {/* Main Content */}
            <main className="max-w-screen-2xl mx-auto px-6 lg:px-10 py-6">
                {children}
            </main>
        </div>
    )
}
