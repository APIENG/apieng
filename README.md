# API Energy Metrics Dashboard

A web application that measures and tracks API energy consumption metrics, powered by AI-generated insights.

## Features

- **API Metrics Measurement**: Track response time, request/response size, and energy consumption for any API endpoint
- **AI-Powered Analysis**: Gemini AI generates contextual metrics and performance insights
- **User Authentication**: Secure signup/login with session management
- **Dashboard**: Visual analytics of measured endpoints
- **CSV Export**: Download metrics in CSV format
- **API Keys**: Generate and manage API keys for programmatic access

## Architecture

- **Backend**: Go (Gorilla Mux router, SQLite)
- **Frontend**: HTML/CSS/JavaScript templates
- **AI Integration**: Gemini API for metrics analysis
- **Auth**: Session tokens + JWT support

## Quick Start

### Prerequisites

1. **API Key and RSA Keys**
   - Place `app.rsa` and `app.rsa.pub` in the project root
   - If missing, the server will fail at startup with key errors

2. **Environment Variables**
   - Create a `.env` file (see `.env.example`)
   - Set `GEMINI_API_KEY` for AI features; without it, AI analysis silently returns blank responses

### Installation

```bash
# Clone the repository
git clone <repo>
cd apieng

# Install dependencies
go mod tidy

# Create RSA keys (if missing)
openssl genrsa -out app.rsa 2048
openssl rsa -in app.rsa -pubout -out app.rsa.pub

# Create .env with your API key
echo "GEMINI_API_KEY=<your-key-here>" > .env
```

### Running the Server

```bash
# From cmd/ directory
go run server.go

# Server starts at http://localhost:8080
```

## API Endpoints

### Authentication
- `POST /login` – User login (form-based)
- `GET /logout` – User logout
- `POST /users` – User signup (form-based)
- `POST /generate` – Generate API key (requires session)

### Metrics (Authenticated)
- `GET /metrics` – View all user metrics (web UI)
- `GET /metrics/{id}` – View single metric details
- `POST /measure` – Measure an API endpoint (form-based, requires session)
- `GET /metrics/export/download` – Export all metrics as CSV
- `GET /metrics/export/download/{id}` – Export single metric as CSV

### API (API Key Auth)
- `GET /api/metrics` – Return user metrics as JSON
- `POST /api/measure` – Measure endpoint and return result as JSON

### Dashboard
- `GET /dashboard` – Analytics dashboard (requires session)

## Security Notes

⚠️ **Known Limitations:**
- Session tokens are not validated against a store — any non-empty session_token cookie is accepted
- The energy consumption model is hardcoded (not real measurement)
- Cookies lack Secure flag for HTTPS-only transmission

## Database

SQLite database (`metrics.db`) with tables:
- `users` – User accounts
- `metrics` – API endpoint measurements

## Development

To run tests (after proper test structure is added):
```bash
go test ./...
```

## Contributing

Contributions welcome. Please ensure:
- All endpoints are properly authenticated
- RSA keys are in place before running
- `.env` is set up with required API keys
