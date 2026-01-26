import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { ThemeProvider } from './contexts/ThemeContext'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import BrandSetup from './pages/BrandSetup'
import RunAnalysis from './pages/RunAnalysis'
import Login from './pages/Login'
import Signup from './pages/Signup'

function App() {
  return (
    <ThemeProvider>
      <Router>
        <Routes>
          {/* Auth routes (no layout) */}
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} />

          {/* App routes (with layout) */}
          <Route path="/" element={<Layout><Dashboard /></Layout>} />
          <Route path="/brands" element={<Layout><BrandSetup /></Layout>} />
          <Route path="/analysis" element={<Layout><RunAnalysis /></Layout>} />
        </Routes>
      </Router>
    </ThemeProvider>
  )
}

export default App
