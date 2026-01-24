// API client for AI Visibility Tracker
const API_BASE = '/api/v1';

// Get auth token from localStorage
function getAuthToken() {
    return localStorage.getItem('token');
}

// Helper function for API calls with error handling
async function apiCall(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;
    const token = getAuthToken();

    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    // Add auth token if available
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(url, {
        headers,
        ...options,
    });

    // Handle non-JSON responses (HTML error pages, etc.)
    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
        if (!response.ok) {
            throw {
                status: response.status,
                message: `Server error: ${response.status} ${response.statusText}`,
            };
        }
        // For non-JSON success responses (like CSV downloads), return text
        return response.text();
    }

    const data = await response.json();

    if (!response.ok) {
        throw {
            status: response.status,
            ...data,
        };
    }

    return data;
}

// ============================================
// Auth APIs
// ============================================

export async function login(email, password) {
    return apiCall('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    });
}

export async function signup(email, password, name) {
    return apiCall('/auth/signup', {
        method: 'POST',
        body: JSON.stringify({ email, password, name }),
    });
}

export async function getMe() {
    return apiCall('/me');
}

export function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
}

export function isAuthenticated() {
    return !!getAuthToken();
}

export function getCurrentUser() {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
}

// ============================================
// Brand APIs
// ============================================

export async function getBrands() {
    return apiCall('/brands');
}

export async function createBrand(brandData) {
    return apiCall('/brands', {
        method: 'POST',
        body: JSON.stringify(brandData),
    });
}

export async function getBrand(id) {
    return apiCall(`/brands/${id}`);
}

export async function updateBrand(id, brandData) {
    return apiCall(`/brands/${id}`, {
        method: 'PUT',
        body: JSON.stringify(brandData),
    });
}

export async function deleteBrand(id) {
    return apiCall(`/brands/${id}`, {
        method: 'DELETE',
    });
}

export async function updateAlertSettings(id, alertThreshold, scheduleFrequency) {
    return apiCall(`/brands/${id}/alerts`, {
        method: 'PUT',
        body: JSON.stringify({
            alert_threshold: alertThreshold,
            schedule_frequency: scheduleFrequency,
        }),
    });
}

// ============================================
// Competitor APIs
// ============================================

export async function getCompetitors(brandId) {
    return apiCall(`/brands/${brandId}/competitors`);
}

export async function addCompetitor(brandId, name) {
    return apiCall(`/brands/${brandId}/competitors`, {
        method: 'POST',
        body: JSON.stringify({ name }),
    });
}

export async function removeCompetitor(brandId, competitorId) {
    return apiCall(`/brands/${brandId}/competitors/${competitorId}`, {
        method: 'DELETE',
    });
}

// ============================================
// Alias APIs
// ============================================

export async function getAliases(brandId) {
    return apiCall(`/brands/${brandId}/aliases`);
}

export async function addAlias(brandId, alias) {
    return apiCall(`/brands/${brandId}/aliases`, {
        method: 'POST',
        body: JSON.stringify({ alias }),
    });
}

export async function removeAlias(brandId, aliasId) {
    return apiCall(`/brands/${brandId}/aliases/${aliasId}`, {
        method: 'DELETE',
    });
}

// ============================================
// Prompt APIs
// ============================================

export async function getPrompts() {
    return apiCall('/prompts');
}

export async function createPrompt(category, template, description = '') {
    return apiCall('/prompts', {
        method: 'POST',
        body: JSON.stringify({ category, template, description }),
    });
}

export async function deletePrompt(id) {
    return apiCall(`/prompts/${id}`, {
        method: 'DELETE',
    });
}

export async function updatePrompt(id, category, template, description = '') {
    return apiCall(`/prompts/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ category, template, description }),
    });
}

// ============================================
// Analysis APIs (with rate limiting protection)
// ============================================

// Check if we can run analysis before attempting
export async function getAnalysisStatus() {
    return apiCall('/analysis/status');
}

// Run analysis - the backend enforces rate limiting
export async function runAnalysis(brandId, promptIds = []) {
    return apiCall('/analysis/run', {
        method: 'POST',
        body: JSON.stringify({
            brand_id: brandId,
            prompt_ids: promptIds,
        }),
    });
}

export async function getAnalysisResults(brandId) {
    return apiCall(`/analysis/results?brand_id=${brandId}`);
}

export async function getAnalysisResult(id) {
    return apiCall(`/analysis/results/${id}`);
}

// ============================================
// Metrics APIs
// ============================================

export async function getMetrics(brandId) {
    return apiCall(`/metrics?brand_id=${brandId}`);
}

export async function getDashboardData(brandId) {
    return apiCall(`/metrics/dashboard?brand_id=${brandId}`);
}

// ============================================
// Export APIs
// ============================================

export async function exportCSV(brandId) {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_BASE}/export/csv?brand_id=${brandId}`, {
        headers: {
            'Authorization': `Bearer ${token}`,
        },
    });
    if (!response.ok) {
        throw new Error('Export failed');
    }
    // Get filename from header or use default
    const contentDisposition = response.headers.get('Content-Disposition');
    const filename = contentDisposition?.match(/filename=(.+)/)?.[1] || `visibility_report.csv`;

    // Download the file
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
    return true;
}

// ============================================
// Health Check
// ============================================

export async function healthCheck() {
    return apiCall('/health'.replace('/api/v1', ''));
}
