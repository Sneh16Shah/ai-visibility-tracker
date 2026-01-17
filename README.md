# AI Visibility Tracker

A portfolio-grade AI analytics platform that measures how brands appear in AI-generated responses. Built to demonstrate system design, AI integration, and full-stack development skills.

## 🎯 Problem Statement

As AI tools increasingly replace traditional search, brands lack visibility into:
- Whether AI systems mention them
- How often they appear vs competitors
- The sentiment/context of AI recommendations

This project explores how such visibility can be measured and visualized.

## 🏗️ Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│   React + Vite  │────▶│   Go (Gin) API  │────▶│   PostgreSQL    │
│   Dashboard     │     │   Backend       │     │   Database      │
│                 │     │                 │     │                 │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │                 │
                        │   AI Provider   │
                        │  (OpenAI/LLM)   │
                        │                 │
                        └─────────────────┘
```

## ⚡ Tech Stack

| Layer | Technology |
|-------|------------|
| **Frontend** | React, Vite, Tailwind CSS, Recharts |
| **Backend** | Go, Gin Framework |
| **Database** | PostgreSQL |
| **AI Layer** | OpenAI API / Ollama (local LLM) |

## 🚀 Features

- **Brand Configuration** - Set up brands, aliases, and competitors
- **Prompt Templates** - Predefined prompts for AI visibility analysis
- **Manual Analysis** - One-click AI response generation
- **Mention Detection** - Automatic brand/competitor detection
- **Sentiment Analysis** - Positive/neutral/negative classification
- **Analytics Dashboard** - Visibility scores, citation share, trends

## 📊 Key Metrics

- **Visibility Score** - Overall brand visibility in AI responses
- **Citation Share** - Percentage of responses mentioning your brand
- **Mention Frequency** - Count of brand mentions over time
- **Sentiment Breakdown** - Distribution of positive/neutral/negative mentions

## 🛠️ Getting Started

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+

### Backend Setup
```bash
cd backend
go mod download
go run main.go
```

### Frontend Setup
```bash
cd frontend
npm install
npm run dev
```

### Environment Variables
```
# Backend
PORT=8080
DATABASE_URL=postgres://localhost:5432/ai_visibility_tracker
OPENAI_API_KEY=your-key-here

# Frontend (via Vite proxy)
# API calls automatically proxy to localhost:8080
```

### 🐳 Docker Setup (Recommended)

Run the entire stack with a single command:

**1. Create a `.env` file in the project root:**
```bash
GEMINI_API_KEY=your-gemini-api-key
GROQ_API_KEY=your-groq-api-key
DB_PASSWORD=your-secure-password
JWT_SECRET=your-jwt-secret
```

**2. Start all services:**
```bash
docker-compose up -d
```

This starts:
- **MySQL** on port `3306`
- **Backend API** on port `8080`
- **Frontend** on port `80`

**3. Access the app:**
- Frontend: http://localhost
- Backend API: http://localhost:8080

**4. View logs:**
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
```

**5. Rebuild after code changes:**
```bash
# Rebuild and restart a specific service
docker-compose up -d --build backend
docker-compose up -d --build frontend

# Rebuild all services
docker-compose up -d --build
```

**6. Stop and cleanup:**
```bash
# Stop services (keeps data)
docker-compose down

# Stop and remove volumes (removes database data)
docker-compose down -v
```

## 📁 Project Structure

```
ai-visibility-tracker/
├── backend/
│   ├── main.go              # Entry point
│   ├── config/              # Configuration
│   ├── routes/              # API routes
│   ├── controllers/         # HTTP handlers
│   ├── services/            # Business logic
│   ├── models/              # Data structures
│   ├── db/                  # Database layer
│   └── ai/                  # AI provider abstraction
├── frontend/
│   ├── src/
│   │   ├── components/      # Reusable UI components
│   │   ├── pages/           # Route pages
│   │   └── api/             # API client
│   └── public/
├── prd.md                   # Product Requirements
└── project-execution-guide.md
```

## 🎨 Design Decisions

1. **Monolithic Backend** - Simpler deployment for a portfolio project
2. **Manual Execution** - No schedulers needed, conserves API credits
3. **Permanent Storage** - All AI responses stored for analysis
4. **Provider Abstraction** - Easy to swap OpenAI for local LLM

## 📝 License

MIT License - Free to use for personal and educational purposes.

---

Built by [Sneh Shah](https://github.com/Sneh16Shah)
