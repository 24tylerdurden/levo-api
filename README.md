# Levo API

A simple Go API server with SQLite database support, designed for Levo.ai's automated pen-testing tool.

## Features

- Go-based HTTP server with Gin framework
- SQLite database with migrations
- Docker containerization
- Health check endpoint
- Graceful shutdown handling

## Database Schema

The application includes the following tables:
- `applications` - Stores application information
- `services` - Stores service information (belongs to applications)
- `schema_versions` - Stores versioned API schemas for applications/services

## Quick Start with Docker

### Prerequisites
- Docker and Docker Compose installed

### Running the Application

1. **Build and start the container:**
   ```bash
   docker-compose up -d
   ```

2. **Check if the service is running:**
   ```bash
   docker-compose ps
   ```

3. **Test the health endpoint:**
   ```bash
   curl http://localhost:8080/health
   ```

   Expected response:
   ```json
   {
     "database": "connected",
     "status": "healthy"
   }
   ```

4. **View logs:**
   ```bash
   docker-compose logs -f
   ```

5. **Stop the service:**
   ```bash
   docker-compose down
   ```

## Project Structure

```
levo-api/
├── cmd/server/           # Main application entry point
├── internal/database/     # Database connection and migration logic
├── pkg/config/           # Configuration management
├── migrations/           # SQL migration files
├── data/                 # SQLite database files (created at runtime)
├── storage/              # File storage directory
├── Dockerfile            # Docker build configuration
├── docker-compose.yml    # Docker Compose configuration
└── go.mod               # Go module dependencies
```

## Environment Variables

The application supports the following environment variables:

- `LEVO_DB_PATH` - Path to SQLite database file (default: `/app/data/levo.db`)
- `LEVO_STORAGE_PATH` - Path to file storage directory (default: `/app/storage`)
- `LEVO_MIGRATIONS_PATH` - Path to migration files (default: `/app/migrations`)
- `LEVO_PORT` - Server port (default: `8080`)

## Development

### Running Locally (without Docker)

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Run the application:**
   ```bash
   go run cmd/server/main.go
   ```

3. **Test the health endpoint:**
   ```bash
   curl http://localhost:8080/health
   ```

## Database Migrations

The application automatically runs database migrations on startup. Migration files are located in the `migrations/` directory:

- `001_initial_schema.up.sql` - Creates initial tables
- `001_initial_schema.down.sql` - Drops tables (for rollback)

