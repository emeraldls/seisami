# Seisami

**Voice-driven task management for productivity without friction.**

Record tasks with a hotkey. AI transcribes and organizes them automatically. Collaborate in real-time. No account required. 100% open source.

## Quick Start

```bash
# Development
wails dev

# Build
wails build

# Desktop app available at: build/bin/
```

## What is Seisami?

Seisami is a **voice-first task management platform** that eliminates the friction between thinking and organizing:

- **Voice Recording** - Press FN anywhere to capture tasks
- **AI Processing** - GPT-4.1 understands context and extracts intent
- **Kanban Boards** - Automatic task categorization and drag-and-drop organization
- **Real-Time Collaboration** - Share boards with team, sync instantly
- **Offline-First** - Works completely offline, no account needed

## Project Structure

```
seisami/
├── app/                    # Desktop application (Go + React)
│   ├── frontend/          # React 19, TypeScript, Tailwind
│   ├── internal/          # Go logic, AI integration
│   └── main.go            # Wails entry point
├── server/                # Central collaboration server (Go)
│   ├── central/           # Auth & handler
│   └── centraldb/         # PostgreSQL operations
└── web/                   # Marketing website (TanStack Start)
    └── src/               # React components, pages
```

## Development

### Prerequisites
- Go 1.25+
- Node.js 18+
- npm or pnpm

### Desktop App (`/app`)

```bash
cd app
go mod download
npm install
wails dev              # Hot reload development
wails build            # Production binary
```

**Frontend**: React 19 + TypeScript + Tailwind CSS
**Backend**: Go + Wails framework
**Database**: SQLite (local)
**AI**: OpenAI API (configurable)

### Server (`/server`)

```bash
cd server
go run main.go
```

Runs on:
- HTTP: `:8080` (authentication)
- TCP: `:2121` (real-time sync)

**Database**: PostgreSQL 15+
**Protocol**: Custom TCP with JSON messages

### Website (`/web`)

```bash
cd web
npm install
npm run dev            # http://localhost:3000
npm run build          # Production build
```

**Framework**: TanStack Start (full-stack React)
**Styling**: Tailwind CSS v4

## 🔧 Configuration

### Desktop App
Edit `app/wails.json`:
```json
{
  "appname": "Seisami",
  "frontend": {
    "installprefix": "./frontend"
  }
}
```

### Server
Environment variables:
```bash
DATABASE_URL=postgresql://user:pass@localhost/seisami
HTTP_ADDR=:8080
TCP_ADDR=:2121
JWT_SECRET=your-secret-key
```

### Website
Environment variables:
```bash
VITE_BASE_URL=https://seisami.com
VITE_SITE_NAME=Seisami
```

## 📊 Database Schema

### Desktop (SQLite)
```sql
boards, columns, cards          -- Core kanban structure
transcriptions                  -- Audio + AI processing
settings                        -- User configuration
```

### Server (PostgreSQL)
```sql
users, sessions                 -- Authentication
rooms, room_members             -- Collaboration
```

## Privacy & Security

- **Local-First**: All data stored locally by default
- **No Tracking**: Zero telemetry
- **Optional Cloud**: Connect to server for team collaboration
- **Open Source**: Full code transparency for security audit
- **Encryption**: Optional AES-256 for sensitive data
- **No Account Required**: Works out-of-the-box

## Project Goals

1. **Zero Friction** - Recording tasks should be faster than typing
2. **Privacy First** - Users own their data
3. **Open & Transparent** - No vendor lock-in
4. **Beautiful Design** - Simple, minimal, professional
5. **Team Ready** - Optional collaboration without complexity