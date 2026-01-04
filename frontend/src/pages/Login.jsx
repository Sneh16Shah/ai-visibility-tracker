import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import * as api from '../api/client'

export default function Login() {
    const navigate = useNavigate()
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(false)

    const handleSubmit = async (e) => {
        e.preventDefault()
        setError('')
        setLoading(true)

        try {
            const response = await api.login(email, password)

            // Store token and user in localStorage
            localStorage.setItem('token', response.token)
            localStorage.setItem('user', JSON.stringify(response.user))

            // Redirect to dashboard
            navigate('/')
        } catch (err) {
            setError(err.error || 'Login failed. Please check your credentials.')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="min-h-screen bg-[var(--background)] flex items-center justify-center p-4">
            <div className="w-full max-w-md">
                {/* Logo */}
                <div className="text-center mb-8">
                    <h1 className="text-3xl font-bold bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 bg-clip-text text-transparent">
                        AI Visibility Tracker
                    </h1>
                    <p className="text-[var(--text-muted)] mt-2">Sign in to your account</p>
                </div>

                {/* Login Form */}
                <div className="bg-[var(--surface)] rounded-2xl p-8 border border-[var(--surface-light)]">
                    <form onSubmit={handleSubmit} className="space-y-6">
                        {error && (
                            <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-3 text-red-400 text-sm">
                                {error}
                            </div>
                        )}

                        <div>
                            <label className="block text-sm font-medium text-[var(--text-muted)] mb-2">
                                Email
                            </label>
                            <input
                                type="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                required
                                className="w-full px-4 py-3 bg-[var(--background)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)] transition-colors"
                                placeholder="you@example.com"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-[var(--text-muted)] mb-2">
                                Password
                            </label>
                            <input
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                required
                                className="w-full px-4 py-3 bg-[var(--background)] border border-[var(--surface-light)] rounded-lg text-[var(--text)] focus:outline-none focus:border-[var(--primary)] transition-colors"
                                placeholder="••••••••"
                            />
                        </div>

                        <button
                            type="submit"
                            disabled={loading}
                            className="w-full py-3 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white rounded-lg font-medium transition-all duration-200 disabled:opacity-50"
                        >
                            {loading ? 'Signing in...' : 'Sign In'}
                        </button>
                    </form>

                    <div className="mt-6 text-center text-sm text-[var(--text-muted)]">
                        Don't have an account?{' '}
                        <Link to="/signup" className="text-[var(--primary)] hover:underline">
                            Sign up
                        </Link>
                    </div>

                    {/* Demo credentials */}
                    <div className="mt-6 p-4 bg-[var(--background)] rounded-lg">
                        <p className="text-xs text-[var(--text-muted)] text-center">
                            Demo: <span className="text-[var(--text)]">demo@example.com</span> / <span className="text-[var(--text)]">demo123</span>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    )
}
