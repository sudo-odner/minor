# Agent Instructions

This document provides guidance for AI agents working in this repository.

## Project Overview

This is a real-time communication platform with a React frontend and Go microservices backend.

```
minor-kursach/
├── frontend/              # React 19 application
├── backend/
│   ├── services/         # Go microservices (auth_service, user_service, etc.)
│   └── shared/           # Shared Go modules
└── deploy/               # Docker deployment configs
```

## Build Commands

### Frontend (React)
```bash
cd frontend/
npm install              # Install dependencies
npm start                # Start dev server (http://localhost:3000)
npm run build            # Production build
```

### Backend (Go)
```bash
cd backend/
go work sync             # Sync workspace dependencies
go build ./...           # Build all services
go mod download          # Download dependencies
go mod tidy             # Clean up go.mod/go.sum
```

### Running a Single Service
```bash
cd backend/services/auth_service/
go run cmd/auth/main.go  # Run auth service
```

### Database Migrations (user_service)
```bash
cd backend/services/user_service/
make migrate-up              # Apply all migrations
make migrate-down            # Rollback last migration
make migrate-force version=N # Force to specific version
make migrate-version        # Show current version
make migrate-create name=xxx # Create new migration
```
**Note:** Requires `migrate` CLI tool installed.

## Test Commands

### Frontend Tests (Jest + React Testing Library)
```bash
cd frontend/
npm test                              # Run all tests (watch mode)
npm test -- --watchAll=false          # Run once (CI mode)
npm test -- --testPathPattern=App     # Run specific test file
npm test -- -t "test name"            # Run test by name pattern
npm test -- --coverage                # Generate coverage report
```

### Backend Tests (Go)
```bash
cd backend/
go test ./...              # Run all tests
go test ./services/auth_service/...  # Test specific service
go test -v ./...          # Verbose output
go test -run TestName     # Run specific test by name
```

**Note:** Most backend services currently have no test files.

## Code Style

### Go (Backend)

**Formatting:**
- Use standard `go fmt` formatting (tabs for indentation)
- Run `go fmt ./...` before committing

**Naming Conventions:**
- Package names: lowercase, single words (e.g., `service`, `repository`)
- Struct names: PascalCase (e.g., `AuthorizationService`)
- Variable names: camelCase (e.g., `authRepository`)
- Constants: camelCase or SCREAMING_SNAKE_CASE for config constants
- Files: lowercase with underscores (e.g., `auth.go`, `postgres.go`)

**Directory Structure (per service):**
```
internal/
├── app/              # Application setup
├── config/           # Configuration loading (YAML/env)
├── http-server/     # HTTP handlers and middleware
│   ├── handler/      # Request handlers
│   └── middleware/   # HTTP middleware (CORS, JWT, etc.)
├── lib/              # Shared libraries (logger, JWT, response)
├── model/            # Data models
├── repository/       # Database access layer
│   └── postgres/     # PostgreSQL implementations
└── service/         # Business logic layer
```

**Error Handling Pattern:**
```go
const op = "service.auth.Login"

func (s *Service) DoSomething() error {
    err := doSomething()
    if err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }
    return nil
}
```

**Logging:**
- Use `go.uber.org/zap` for structured logging
- Always include operation context: `log.With(zap.String("op", op))`

**Imports:**
- Standard library first, then third-party, then internal
- Use aliases for disambiguation (e.g., `authHandler "..."`)
- Group with blank lines between groups

**Dependencies:**
- Chi v5 (`github.com/go-chi/chi/v5`) for HTTP routing
- pgx/v5 (`github.com/jackc/pgx/v5`) for PostgreSQL
- zap for logging
- cleanenv for config loading

### JavaScript/React (Frontend)

**Formatting:**
- No Prettier config; follow ESLint defaults
- Use semicolons at end of statements
- Use double quotes for strings

**Naming Conventions:**
- Components: PascalCase (e.g., `VoiceChannel.jsx`)
- Functions/Variables: camelCase
- Constants: SCREAMING_SNAKE_CASE

**Component Patterns:**
```jsx
import { useState, useEffect } from 'react';

const VoiceChannel = ({ channelId, userId }) => {
  const [isActive, setIsActive] = useState(false);

  useEffect(() => {
    // cleanup
    return () => {};
  }, []);

  return <div className="voice-channel">{/* ... */}</div>;
};

export default VoiceChannel;
```

**Imports:**
- React imports first
- Then external libraries
- Then internal components/utils
- Then CSS/assets

**Testing:**
- Use `@testing-library/react` and `@testing-library/user-event`
- Follow patterns in `src/App.test.js`

## Common Patterns

### Creating a New Service
1. Create directory under `backend/services/<service_name>/`
2. Initialize Go module: `go mod init github.com/sudo-odner/minor/backend/services/<service_name>`
3. Create `internal/` structure with app, config, handlers, etc.
4. Add to `backend/go.work`

### Creating a New API Endpoint
1. Add handler method in appropriate handler file
2. Register route in `main.go` using chi router
3. Follow the pattern: handler -> service -> repository

### Environment Variables
- Store in `.env` files (not committed)
- Use `cleanenv` for YAML/env loading
- Required vars documented in config files

## Architecture Notes

- **API Gateway**: Routes requests to appropriate services
- **Auth Service**: JWT-based authentication
- **User Service**: User management and relationships
- **NATS**: Inter-service communication
- **PostgreSQL**: Primary database (user_service, auth_service)
- **Traefik**: Reverse proxy and load balancer

## Configuration Files

| File | Purpose |
|------|---------|
| `backend/go.work` | Go workspace configuration |
| `frontend/package.json` | Frontend dependencies and scripts |
| `deploy/docker-compose.yaml` | Full stack deployment |
| `.env` (per service) | Environment variables |
