# SynapseChain

> A trust layer for data — AI + Humans collaboratively label data with accuracy, explainability, and verifiability.

## Architecture

```
Client → API Gateway (Go :8080)
              ↓
       AI Labeling Service (Python :8081)
              ↓
        Routing Engine (confidence-driven)
              ↓
       Human Validation UI (React :3000)
              ↓
        Consensus Engine
              ↓
        PostgreSQL → Output API
```

## Tech Stack

| Layer    | Technology       |
|----------|-----------------|
| Backend  | Go (gorilla/mux)|
| AI       | Python (FastAPI) |
| Frontend | React TypeScript |
| Database | PostgreSQL 16    |
| Infra    | Docker Compose   |

## Quick Start

### Prerequisites
- Docker & Docker Compose

### Run everything
```bash
docker compose up --build
```

Services:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **AI Service**: http://localhost:8081
- **PostgreSQL**: localhost:5432

### Development (without Docker)

**Database:**
```bash
# Start PostgreSQL and run migrations
psql -U synapse -d synapsechain -f db/migrations/001_init.sql
```

**AI Service:**
```bash
cd ai-service
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt
uvicorn app.main:app --port 8081 --reload
```

**Backend:**
```bash
cd backend
cp ../.env.example ../.env
go run ./cmd/api
```

**Frontend:**
```bash
cd frontend
npm install
npm start
```

## API Reference

### Data Ingestion

**Upload event/content data:**
```
POST /api/v1/data/upload
Content-Type: application/json

{
  "type": "event",
  "data": {
    "speed": 130,
    "deceleration": -0.6,
    "angular_velocity": 0.2
  }
}

→ 201 { "data_id": "uuid", "status": "received" }
```

**Upload binary file:**
```
POST /api/v1/data/upload/file
Content-Type: multipart/form-data

file: <binary>
type: "content"

→ 201 { "data_id": "uuid", "status": "received" }
```

### Labels

**Get final label:**
```
GET /api/v1/label/{data_id}

→ 200 {
  "final_label": { "event_type": "hard_braking", "risk": "medium" },
  "confidence": 0.92,
  "source": ["AI", "Human"]
}
```

**Get AI label:**
```
GET /api/v1/label/{data_id}/ai
```

### Human Validation

**List pending validations:**
```
GET /api/v1/validation/pending

→ 200 [{ "data_id": "...", "labels": {...}, "confidence": 0.76 }]
```

**Submit validation:**
```
POST /api/v1/validation/submit
{
  "data_id": "uuid",
  "validator_id": "user_123",
  "action": "accept|modify|reject",
  "labels": {},
  "notes": "optional"
}
```

## How It Works

1. **Upload**: Data (sensor JSON, video, text) is ingested via API
2. **AI Labels**: Python service classifies data and assigns confidence scores
3. **Routing**: If confidence < threshold OR high-risk → routed to human review
4. **Validation**: Humans accept, modify, or reject AI labels via dashboard
5. **Consensus**: Final label is produced (human label takes priority over AI)
6. **Output**: Labeled data is available via API with confidence + source info

## Project Structure

```
SynapseChain/
├── backend/                 # Go API server
│   ├── cmd/api/main.go      # Entry point
│   └── internal/
│       ├── config/           # Environment config
│       ├── db/               # Database connection
│       ├── handler/          # HTTP handlers + routes
│       ├── models/           # Data models
│       ├── repository/       # Database queries
│       └── service/          # Business logic (AI client, routing, consensus)
├── ai-service/              # Python AI labeling
│   └── app/
│       ├── main.py           # FastAPI server
│       └── labeler.py        # Event & content classifiers
├── frontend/                # React TypeScript UI
│   └── src/
│       ├── components/       # Dashboard, UploadForm, LabelLookup
│       └── api.ts            # API client
├── db/migrations/           # SQL schema
├── docker-compose.yml       # Full stack orchestration
└── .env.example             # Environment template
```

## License

See [LICENSE](LICENSE).
