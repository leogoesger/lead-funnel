# Deployment Guide

This directory contains Docker configuration for the Lead Funnel application.

## Prerequisites

- Docker and Docker Compose installed
- Environment variables configured (see below)

## Quick Start

1. **Copy and configure environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your actual values
   ```

2. **Build and start services:**
   ```bash
   docker-compose up -d
   ```

3. **Check service status:**
   ```bash
   docker-compose ps
   ```

4. **View logs:**
   ```bash
   docker-compose logs -f api
   docker-compose logs -f postgres
   ```

## Services

### PostgreSQL 18.4
- **Container:** lead-funnel-postgres
- **Port:** 5432
- **Database:** lead_funnel
- **User:** lead_funnel_user
- **Password:** From `DB_PASSWORD` in .env

### API Server
- **Container:** lead-funnel-api
- **Port:** 8080
- **Health Check:** GET /health

### Database Migrations
- **Container:** lead-funnel-migrate
- **Purpose:** Automatically runs migrations on startup

## Environment Variables

Create a `.env` file in this directory with:

```
DB_PASSWORD=your_secure_password
ENVIRONMENT=development
LOG_LEVEL=info
OPENAI_API_KEY=sk-...
TWILIO_ACCOUNT_SID=AC...
TWILIO_AUTH_TOKEN=...
TWILIO_PHONE_NUMBER=+1...
```

## Common Commands

**Start services:**
```bash
docker-compose up -d
```

**Stop services:**
```bash
docker-compose down
```

**Remove data (clean slate):**
```bash
docker-compose down -v
```

**Run migrations manually:**
```bash
docker-compose up migrate
```

**Access PostgreSQL CLI:**
```bash
docker-compose exec postgres psql -U lead_funnel_user -d lead_funnel
```

**View API logs:**
```bash
docker-compose logs -f api
```

**Rebuild images:**
```bash
docker-compose build --no-cache
```

## Troubleshooting

**Migrations fail:**
- Check `docker-compose logs migrate`
- Ensure migrations are in `../migrations` directory

**API can't connect to database:**
- Wait for PostgreSQL health check to pass
- Check `DATABASE_URL` environment variable
- Verify `DB_PASSWORD` matches

**Port already in use:**
- Change ports in docker-compose.yml or:
  ```bash
  docker-compose down  # Stop conflicting service
  ```

## Development Notes

- PostgreSQL data is persisted in the `postgres_data` volume
- Migrations run automatically via the `migrate` service
- API health check endpoint: `http://localhost:8080/health`
- Database connection string: `postgres://lead_funnel_user:PASSWORD@postgres:5432/lead_funnel`
