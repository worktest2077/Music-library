# Music Library API

## Requirements
- Go 1.x
- PostgreSQL
- Swagger

## Setup
1. Clone repository
2. Copy `.env.example` to `.env` and configure
3. Run: `go mod download`
4. Start PostgreSQL
5. Run: `swag init`
6. Start mock API: `go run mock_server/main.go`
7. Start app: `go run main.go