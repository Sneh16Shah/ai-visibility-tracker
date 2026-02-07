// Available AI models for comparison
// Backend handles the actual API calls via OpenRouter and Groq

export const AI_MODELS = [
    { id: 'google/gemma-3-27b-it:free', name: 'Gemma 3 27B', provider: 'Google', color: '#4285f4' },
    { id: 'meta-llama/llama-3.3-70b-instruct:free', name: 'Llama 3.3 70B', provider: 'Meta', color: '#0668e1' },
    { id: 'qwen/qwen3-coder:free', name: 'Qwen3 Coder', provider: 'Qwen', color: '#6366f1' },
    { id: 'tngtech/deepseek-r1t2-chimera:free', name: 'DeepSeek Chimera', provider: 'TNG', color: '#00d4aa' },
    { id: 'groq', name: 'Groq Llama 3.3', provider: 'Groq', color: '#f55036' },
]
