.PHONY: run docker-up docker-down docker-logs docker-build docker-ps docker-clean

# Local development
run: 
	go run ./cmd/api

# Docker Compose commands
docker-build:
	cd deploy && docker compose -p lead-funnel build

docker-up:
	cd deploy && docker compose -p lead-funnel up -d --build

docker-down:
	cd deploy && docker compose -p lead-funnel down

docker-logs:
	cd deploy && docker compose -p lead-funnel logs -f api

docker-logs-db:
	cd deploy && docker compose -p lead-funnel logs -f postgres

docker-ps:
	cd deploy && docker compose -p lead-funnel 	ps

docker-clean:
	cd deploy && docker compose -p lead-funnel down -v

docker-migrate:
	cd deploy && docker compose -p lead-funnel up migrate

# Restart all services
docker-restart: docker-down docker-up