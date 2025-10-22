## 1. Project Summary

**Seisami** is a comprehensive audio-driven task management platform with three distinct components:

### 1.1 Desktop Application (`/app`)
A local-first desktop application for audio-driven task management built with Wails. Users record audio via a global hotkey (FN key), transcribe speech using local Whisper models or OpenAI's API, and automatically process transcriptions with AI to create or update tasks on a Kanban-style board. The app features board management, column and card CRUD operations, audio recording with PortAudio, transcription processing, and AI-assisted intent extraction for task automation.

### 1.2 Cloud Collaboration Server (`/server`)
A optional central cloud infrastructure providing real-time collaboration, user authentication, and advanced features. Equipped with PostgreSQL-backed user management, JWT authentication, password reset flows, and WebSocket-based room management for multi-user board synchronization. This server is optional—users can use Seisami entirely offline or connect for collaborative features.

### 1.3 Marketing & Web Application (`/web`)
A future-facing web presence and potentially a web app for accessing Seisami in browser environments. Built with TanStack Start for full-stack React, using modern server-side rendering capabilities. Currently scaffolded with demo routes and ready for expansion.

### Architecture Overview
The system follows a **hybrid local-first with optional cloud synchronization** model:
- **Desktop App**: Local SQLite database, embedded room server for local collaboration, optional connection to central server for cloud features
- **Central Server**: PostgreSQL database for users, authentication, and persistent rooms; WebSocket support for real-time sync across clients
- **Web App**: Future bridge to Seisami data and features via web interface

### Core Tech Stack
**Desktop (`/app`):**
- Backend: Go 1.25+ with Wails v2 framework for cross-platform desktop binaries
- Frontend: React 19 with TypeScript, Zustand for state management, @dnd-kit for drag-and-drop, Tailwind CSS + shadcn/ui for styling
- Database: SQLite with SQLC for type-safe SQL queries
- Audio: PortAudio for cross-platform recording, go-audio/wav for WAV encoding
- AI: OpenAI API (GPT-4.1) for transcription and intent extraction, local Whisper support planned
- Collaboration: Custom TCP-based protocol with buffered reader/writer for low-latency room sync
- Build: Wails build system with embedded frontend assets

**Server (`/server`):**
- Backend: Go 1.25+ with standard library HTTP
- Database: PostgreSQL 15+ with pgx/v5 and SQLC for type-safe queries
- Authentication: JWT-based with bcrypt password hashing
- Collaboration: TCP listener on port 2121 with room management and client broadcasting
- Auth HTTP: Runs on configurable port (default :8080) for signup/signin/password-reset flows

**Web (`/web`):**
- Framework: TanStack Start with TanStack Router for full-stack React
- Styling: Tailwind CSS v4 with @tailwindcss/vite
- Development: Vite with nitro server-side rendering support
- Build: Optimized for both SPA and SSR deployment modes

## 2. Core Directories and Responsibilities

### Desktop Application (`/app`)
- **`/app`**: Root entry point with main.go (Wails application setup) and app.go (core application logic)
- **`/app/frontend`**: React TypeScript frontend built with Vite
  - `src/views/`: Primary application views
    - `board.tsx`: Kanban board view with drag-and-drop card management (1000+ lines, complex)
    - `home.tsx`: Home/dashboard view
    - `transcriptions.tsx`: Historical transcriptions and AI processing results
    - `board-management.tsx`: Board CRUD operations
    - `settings.tsx`: Application settings (transcription method, API keys, etc.)
  - `src/components/`: Reusable UI components
    - `collaboration-panel.tsx`: Real-time collaboration UI
    - `transcription-card.tsx`: Card displaying transcription details
    - `transcription-detail-modal.tsx`: Detailed transcription view with intent/response
    - `onboarding-screen.tsx`: Initial onboarding flow
    - `help-dialog.tsx`: Help and documentation
    - `loading.tsx`: Loading states
    - `error-boundary.tsx`: Error handling wrapper
    - `ui/`: shadcn/ui components (button, card, dialog, dropdown-menu, input, label, etc.)
    - `ui/shadcn-io/kanban/index.tsx`: Custom Kanban board component using @dnd-kit
  - `src/stores/`: Zustand state management
    - `board-store.ts`: Board, column, and card state management with Wails integration
    - `collab-store.ts`: Collaboration server connection state and room management
  - `src/contexts/`: React contexts
    - `sidebar-context.tsx`: Sidebar UI state
  - `src/layouts/`:
    - `app-layout.tsx`: Main application layout wrapper
  - `src/lib/utils.ts`: Utility functions for UI
  - `src/utils/utils.ts`: General utilities
- **`/app/internal`**: Core Go business logic
  - `actions/actions.go`: AI processing logic that analyzes transcriptions with GPT-4.1, extracts intents, and generates structured responses. Uses Go OpenAI SDK with function calling.
  - `repo/`: Data access layer with SQLite operations
    - `interfaces.go`: Repository interface defining all database operations
    - `repo.go`: SQLite implementation of Repository interface
    - `sqlc/`: SQLC-generated type-safe database queries
      - `schema.sql`: Database schema with boards, columns, cards, transcriptions, and settings tables
  - `tools/tools.go`: AI function definitions for tool use (CreateCard, UpdateCard, MoveCard, etc.)
- **`/app/types/types.go`**: Shared Go types for messages and data structures (ColumnEvent, Message, etc.)
- **`/app/portaudio`**: Native audio library bindings
  - `include/portaudio.h`: PortAudio header
  - `lib/darwin_amd64/`, `lib/darwin_arm64/`: Platform-specific bindings
- **`/app/include`**: Native C code
  - `fn_key.c`: FN key hotkey registration for audio recording trigger
  - `microphone.m`: macOS microphone access (Objective-C)
  - `sound.m`: macOS audio playback (Objective-C)
- **`/app/build`**: Build output and platform-specific configurations
  - `darwin/`: macOS build assets
  - `windows/`: Windows build assets with installer scripts

### Collaboration Server (`/server`)
- **`/server/main.go`**: Server entry point handling TCP listener on port 2121 and HTTP auth server
- **`/server/central/`**: Authentication and HTTP endpoints
  - `auth_service.go`: JWT-based authentication with bcrypt hashing, signup/signin/password-reset flows
  - `config.go`: Environment-based configuration loading (DATABASE_URL, JWT_SECRET, HTTP_ADDR, etc.)
  - `handler.go`: HTTP handlers for auth endpoints
- **`/server/centraldb/`**: PostgreSQL database operations
  - `db.go`: Database initialization and connection pooling with pgx/v5
  - `models.go`: Database models (User, Session, ResetToken, etc.)
  - `query.sql.go`: SQLC-generated queries for users and auth
- **`/server/client/`**: TCP client connection management
  - `client.go`: Client state, ID generation, send/receive operations
- **`/server/room/`**: Collaboration room logic
  - `room.go`: Room structure with client tracking and broadcasting
- **`/server/room_manager/`**: Room lifecycle management
  - `room_manager.go`: Thread-safe room creation, joining, leaving, and broadcasting
- **`/server/sqlc/`**: Database schema and SQLC configuration
  - `schema.sql`: PostgreSQL schema for users, sessions, and reset tokens
  - `sqlc.yaml`: SQLC code generation config
- **`/server/types/types.go`**: Message types (Message struct with Action, RoomID, Data fields)

### Web Application (`/web`)
- **`/web/src`**: TanStack Start full-stack application
  - `router.tsx`: Route definitions using TanStack Router
  - `routes/__root.tsx`: Root layout
  - `routes/index.tsx`: Home page
  - `routes/demo/`: Demo routes showing TanStack capabilities
    - `start.api-request.tsx`: API request examples
    - `start.server-funcs.tsx`: Server function examples
    - `start.ssr.full-ssr.tsx`: Full SSR examples
    - `start.ssr.data-only.tsx`: Data-only rendering
    - `start.spa-mode.tsx`: SPA mode examples
  - `components/Header.tsx`: Header component
  - `data/demo.punk-songs.ts`: Demo data
- **`/web`**: Build configuration
  - `vite.config.ts`: Vite configuration with TanStack plugins, Tailwind CSS, React, and Nitro support
  - `tsconfig.json`: TypeScript configuration
  - `package.json`: Dependencies (TanStack Start/Router, React 19, Tailwind CSS v4, Vite)

## 3. Design Principles

### Local-First Architecture
Seisami prioritizes offline functionality and local data storage with SQLite as the primary database. The desktop app can operate entirely offline, with optional cloud features disabled. Users can sync data to the central server when connected without losing local data.

### AI-Assisted Voice-Driven Productivity
The core UX revolves around audio capture → transcription → intent extraction → automated task creation. Users speak naturally, and Seisami intelligently converts speech into structured tasks using GPT-4.1 with function calling. The system understands context (board ID, current time, existing columns) to make smart decisions.

### Event-Driven Architecture
- **Frontend-Backend Communication**: Wails runtime events (EventsOn/EventsEmit) for real-time updates (board changes, transcription results, collaboration events)
- **Collaboration**: TCP-based message protocol with JSON serialization for room synchronization; messages include Action, RoomID, and Data fields
- **Reactive UI**: Zustand stores subscribe to events and automatically update components

### Privacy-First Approach
- Local transcription with Whisper planned as alternative to OpenAI API
- API keys stored locally with encryption options
- Users can opt-out of cloud features entirely
- Configurable settings for transcription method, model paths, API endpoints

### Type Safety & Modularity
- Go generics and interfaces for clear contracts
- TypeScript strict mode for frontend type safety
- SQLC-generated SQL queries prevent SQL injection and ensure compile-time type checking
- Clear separation of concerns: UI (components), state management (stores), business logic (actions), data access (repo)

### Cross-Platform Compatibility
- Wails builds target macOS and Windows with native code (C/Objective-C) for platform-specific features (FN key, microphone access, audio playback)
- PortAudio handles audio abstraction across platforms
- Build system automates packaging for distribution

## 4. Database Schema Overview

### Desktop App (SQLite)
```
boards
  ├─ id (TEXT PRIMARY KEY)
  ├─ name (TEXT)
  ├─ created_at, updated_at

columns
  ├─ id (TEXT PRIMARY KEY)
  ├─ board_id (FK → boards)
  ├─ name, position
  
cards
  ├─ id (TEXT PRIMARY KEY)
  ├─ column_id (FK → columns)
  ├─ title, description, attachments
  
transcriptions
  ├─ id (TEXT PRIMARY KEY)
  ├─ board_id (FK → boards)
  ├─ transcription, recording_path
  ├─ intent, assistant_response
  
settings
  ├─ id (SINGLE ROW CHECK)
  ├─ transcription_method (cloud|local|custom)
  ├─ whisper_binary_path, whisper_model_path, openai_api_key
```

### Central Server (PostgreSQL)
```
users
  ├─ id (UUID PRIMARY KEY)
  ├─ email (UNIQUE)
  ├─ password_hash (bcrypt)
  ├─ created_at, updated_at

sessions
  ├─ id (UUID PRIMARY KEY)
  ├─ user_id (FK → users)
  ├─ jwt_token
  ├─ expires_at

password_reset_tokens
  ├─ id (UUID PRIMARY KEY)
  ├─ user_id (FK → users)
  ├─ token
  ├─ expires_at
```

## 5. Communication Protocols

### Desktop App ↔ Central Server
- **Authentication**: HTTP POST to `/auth/signup`, `/auth/signin`, `/auth/password-reset`
- **Collaboration**: TCP connection to port 2121
  - `{"action": "create"}` → Server creates room, returns room ID
  - `{"action": "join", "roomId": "..."}` → Join existing room
  - `{"action": "broadcast", "roomId": "...", "data": "..."}` → Send to all clients in room
  - `{"action": "leave", "roomId": "..."}` → Leave room

### Frontend ↔ Go Backend (Wails)
- Wails binds Go methods to TypeScript, enabling direct function calls
- Example: `await CreateBoard("My Board")` calls Go backend
- Runtime events for async updates: `EventsOn("board:id", callback)` for receiving updates

### Frontend State Management
- **Zustand stores** persist state locally with middleware
- **DevTools integration** for debugging
- **Async operations** with error handling and loading states

## 6. Agent Guidelines

### File Organization
- **Desktop Frontend**: New components in `/app/frontend/src/components/`, views in `/app/frontend/src/views/`
- **Desktop Backend**: Business logic in `/app/internal/actions/`, data access in `/app/internal/repo/`, native code in `/app/include/`
- **Server Backend**: Auth logic in `/server/central/`, room logic in `/server/room/`, database in `/server/centraldb/`
- **Web Frontend**: Routes in `/web/src/routes/`, components in `/web/src/components/`, data loaders in `/web/src/routes/`
- **Naming**: kebab-case for filenames (board-store.ts), PascalCase for components (BoardSelector.tsx), snake_case for Go functions (CreateCard)

### Backend Integration (Desktop App)
- **Exposing Methods**: Add public methods to `App` struct in `/app/app.go` to expose via Wails
- **Pattern**: Methods should accept serializable params, return (result, error)
- **Example**: `func (a *App) CreateBoard(name string) (types.Board, error) { ... }`
- **Frontend Call**: `import { CreateBoard } from "../../wailsjs/go/main/App"` then `await CreateBoard("name")`

### Database Operations
1. Update schema in `/app/internal/repo/sqlc/schema.sql` or `/server/sqlc/schema.sql`
2. Run `sqlc generate` in the respective directory to generate type-safe query code
3. Update repository interface in `/app/internal/repo/interfaces.go`
4. Implement new methods in `/app/internal/repo/repo.go`
5. For server: update queries in `/server/centraldb/` after regeneration

### State Management
- **Board State**: Use `useBoardStore` for boards, columns, cards, onboarding status
- **Collaboration State**: Use `useCollaborationStore` for connection status, room ID, server address
- **Async Operations**: Set loading state before async call, clear on completion
- **Error Handling**: Set error state, use toast notifications (via Sonner) for user feedback

### Audio Processing
- **Recording Flow**:
  1. `startRecording()` initializes PortAudio, detects mic
  2. Records WAV to temp file
  3. `stopRecording()` closes stream, returns path
  4. Send to transcription (local or cloud API)
- **Transcription**:
  1. Call `transcribeWithOpenAI(audioPath)` or `transcribeWithLocalWhisper(audioPath)`
  2. Get transcribed text back
  3. Pass to AI processing for intent extraction
- **Error Handling**: Graceful fallback if PortAudio unavailable, clear error messages for user

### AI Integration (Transcription → Intent → Action)
1. **Transcription Processing** (`app.go` → `ProcessTranscription`)
   - Call OpenAI API or local Whisper
   - Store transcription in database
   - Emit UI event with transcription
2. **Intent Extraction** (`internal/actions/actions.go` → `ProcessWithAI`)
   - Build prompt with board context, current time, user's transcription
   - Call GPT-4.1 with available tools as function definitions
   - Parse tool calls (CreateCard, UpdateCard, MoveCard, etc.)
   - Execute tool handlers to modify database
   - Return structured response with intent, actions taken, results
3. **Tool Handlers** (`internal/tools/tools.go`)
   - Each tool (CreateCard, UpdateCard, etc.) has a handler function
   - Tools fetch board context (columns, cards) as needed
   - Return results as JSON
4. **Frontend Updates**
   - Emit events for successful actions
   - Update Zustand stores to reflect changes
   - Show toast notifications with results

### Collaboration Features
1. **Room Creation**: User initiates in UI → calls `CreateCollaborationRoom()` → backend creates room, gets room ID → emit event with room ID
2. **Joining**: User enters room ID → calls `JoinCollaborationRoom(roomId)` → backend connects to server
3. **Broadcasting**: When user edits board → emit internal event → backend serializes change → sends to collaboration room
4. **Receiving**: Backend listens for incoming messages → parses JSON → updates local state → emits UI event
5. **Message Types**: Define in `/app/types/types.go` (and `/server/types/types.go` for server). Include: action, roomId, data fields

### UI Patterns & Components
- **Kanban Board**: Use custom Kanban component in `/app/frontend/src/components/ui/shadcn-io/kanban/` with @dnd-kit
- **Forms**: Use shadcn/ui Button, Input, Dialog, DropdownMenu for consistency
- **Notifications**: Use Sonner toast (`toast.success()`, `toast.error()`) for feedback
- **Loading**: Show LoadingScreen component during initialization, spinners during operations
- **Error Boundaries**: Wrap features in ErrorBoundary to catch React errors gracefully
- **Responsive Design**: Tailwind CSS grid/flex utilities for responsive layouts

### Server Configuration
- **Environment Variables** (set in `.env`):
  - `DATABASE_URL`: PostgreSQL connection string
  - `JWT_SECRET`: Secret for signing JWTs
  - `HTTP_ADDR`: HTTP server address (default `:8080`)
  - `AUTH_JWT_TTL`: JWT expiration (default `24h`)
  - `AUTH_RESET_TOKEN_TTL`: Password reset token expiration (default `1h`)
- **Startup**: Server initializes PostgreSQL pool, creates schema, starts HTTP server, then listens on TCP port 2121
- **Room Management**: Thread-safe with sync.RWMutex; rooms persist during server runtime

### Web Application Development
- **Routes**: Use TanStack Router's file-based routing in `/web/src/routes/`
- **Server Functions**: Use TanStack Start's server functions for backend logic
- **Data Loading**: Use route loaders for prefetching data before rendering
- **Styling**: Tailwind CSS with TanStack Vite plugin for style extraction
- **Building**: Run `npm run build` to generate optimized SPA or SSR bundle

### Avoid Anti-patterns
- ❌ Direct DOM manipulation outside React (use refs sparingly)
- ❌ Global state outside Zustand stores (causes memory leaks, testing issues)
- ❌ Hardcoded API keys or secrets (use environment variables)
- ❌ Unhandled promises or async operations (always catch errors)
- ❌ Missing loading/error states in async flows
- ❌ Mixing business logic with UI components (extract to stores/services)
- ❌ Skipping type safety (use TypeScript strict, Go interfaces)
- ❌ Not handling network failures gracefully (timeout, reconnect logic)
## 7. Context Awareness

Future agents should maintain consistency by:
- **Preserving local-first architecture**: Prefer local processing and storage over cloud dependencies where possible.
- **Respecting privacy**: Always provide local alternatives to cloud services (e.g., local Whisper vs OpenAI).
- **Handling async flows**: Use proper error handling for TCP connections, audio recording, and API calls. Implement timeouts and reconnection logic for network operations.
- **Maintaining type safety**: Leverage Go's and TypeScript's type systems extensively. Use SQLC for database type safety.
- **Following event-driven patterns**: Use Wails runtime events for frontend-backend communication and TCP messages for collaboration.
- **Ensuring cross-platform compatibility**: Test audio and file system operations across macOS and Windows.
- **Preserving AI-assisted UX**: When adding features, consider how they can leverage transcription data for automation.

## 8. Example Agent Behaviors

### Adding a New Board Feature
1. Define schema changes in `/app/internal/repo/sqlc/schema.sql`
2. Run `sqlc generate` to create type-safe queries
3. Add interface method in `/app/internal/repo/interfaces.go`
4. Implement in `/app/internal/repo/repo.go`
5. Add Wails-exposed method in `/app/app.go`
6. Create React hook/store update in `/app/frontend/src/stores/board-store.ts`
7. Add UI component in `/app/frontend/src/components/`
8. Connect to Zustand store in the component
9. Test across macOS and Windows builds

### Extending AI Processing
1. Add new intent type to `StructuredResponse` in `/app/internal/actions/actions.go`
2. Update system prompt in `buildPromptTemplate()` to explain new intent
3. Add new tool handler in `/app/internal/tools/tools.go`
4. Register tool in `registerTools()` method
5. Test with manual transcriptions in UI
6. Update frontend to handle new response types

### Implementing Real-Time Collaboration
1. Define message type in `/app/types/types.go` and `/server/types/types.go`
2. Add broadcast logic in `/app/app.go` to emit on relevant state changes
3. Add receiving logic to parse incoming messages from `collabReader`
4. Update Zustand stores to react to new message types
5. Update `/server/main.go` to handle new action types in `handleConn()`
6. Test with multiple clients connecting to same room

### Setting Up Server Features
1. Define database schema in `/server/sqlc/schema.sql`
2. Run `sqlc generate` in `/server` directory
3. Add queries to `/server/centraldb/`
4. Add HTTP handlers in `/server/central/handler.go`
5. Register routes in auth service
6. Add environment variable support in `/server/central/config.go`
7. Test with HTTP client (curl, Postman)

### Improving Transcription
1. Extend `transcribeWithOpenAI()` or add `transcribeWithLocalWhisper()` in `/app/app.go`
2. Update settings schema to store method preference
3. Create settings panel in `/app/frontend/src/views/settings.tsx`
4. Call appropriate transcription method based on settings
5. Handle transcription errors gracefully with user-friendly messages
6. Emit UI events with transcription results

## 9. Development Workflow

### Desktop App Development
```bash
cd /app
# Install dependencies
go mod download
npm install --prefix frontend

# Run in dev mode with hot reload
wails dev

# Build for release
wails build -p
```

### Server Development
```bash
cd /server
# Set environment variables
export DATABASE_URL="postgres://..."
export JWT_SECRET="your-secret"

# Run
go run main.go

# Test collaboration: nc localhost 2121
# Send JSON: {"action":"create"}\n
```

### Web App Development
```bash
cd /web
npm install
npm run dev  # Starts on port 3000
npm run build
```

### Database Migrations
- **Desktop**: SQLC regeneration only (schema-driven)
  ```bash
  cd /app/internal/repo/sqlc && sqlc generate
  ```
- **Server**: SQLC regeneration
  ```bash
  cd /server/sqlc && sqlc generate
  ```

## 10. Common Tasks & Patterns

### Creating a New Card
```go
// In app.go
func (a *App) CreateCard(columnId, title, description string) (types.Card, error) {
    card, err := a.repository.CreateCard(columnId, title, description)
    if err != nil {
        return types.Card{}, err
    }
    // Broadcast to collaborators if in room
    if a.collabConn != nil {
        a.broadcastCardCreated(card)
    }
    return card, nil
}
```

### Updating Board State
```typescript
// In board-store.ts
async updateCard(cardId: string, title: string, description: string) {
    this.setIsLoading(true);
    try {
        const result = await UpdateCard(cardId, title, description);
        this.setBoards(this.boards.map(b => 
            b.id === this.currentBoard?.id 
                ? { ...b, updatedAt: new Date().toISOString() }
                : b
        ));
        toast.success("Card updated");
    } catch (err) {
        this.setError("Failed to update card");
        toast.error("Update failed");
    } finally {
        this.setIsLoading(false);
    }
}
```

### Handling Collaboration Messages
```go
// In app.go
case "card:moved":
    var event types.CardMovedEvent
    json.Unmarshal(msg.Data, &event)
    // Update local store via Wails event
    runtime.EventsEmit(a.ctx, "card:moved:remote", event)
```

## 11. Performance Considerations

- **Audio Recording**: PortAudio runs in separate goroutine to avoid blocking UI
- **Large Boards**: Implement pagination/lazy loading for boards with 100+ columns/cards
- **Collaboration**: Use buffered channels to prevent message loss during high traffic
- **AI Calls**: Cache commonly used intent patterns; consider request batching
- **Database**: Use indexes on foreign keys (board_id, column_id) for faster queries
- **Frontend**: Memoize expensive computations with useMemo, use React.lazy for code splitting

## 12. Testing Strategy

- **Unit Tests**: Test individual functions in actions, tools, repo (use Go testing package, Jest for JS)
- **Integration Tests**: Test full flows (e.g., transcription → intent → card creation)
- **Collaboration Tests**: Multiple clients connecting, sending messages, state sync
- **UI Tests**: Component snapshots, user interaction flows with React Testing Library
- **Cross-Platform**: Always test on both macOS and Windows builds

## 13. Deployment & Distribution

### Desktop App
- Wails auto-packages as `.dmg` (macOS) and `.exe` installer (Windows)
- Code signing required for distribution on macOS
- Update mechanism: Check release version, prompt user to download

### Server
- Docker container recommended for central server
- PostgreSQL hosted on managed service or self-hosted
- Environment variables for configuration
- Monitor TCP port 2121 for collaboration connections

### Web
- Deploy to Vercel, Netlify, or self-hosted Node/static server
- Environment variables for API endpoints
- CDN for static assets