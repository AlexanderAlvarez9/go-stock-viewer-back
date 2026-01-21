#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}=== Stock Viewer Backend - Local Development ===${NC}"
echo ""

if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    exit 1
fi

DOCKER_COMPOSE=""
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    exit 1
fi

echo -e "${YELLOW}Using: ${DOCKER_COMPOSE}${NC}"

echo -e "${YELLOW}Installing swag for Swagger documentation...${NC}"
go install github.com/swaggo/swag/cmd/swag@latest

echo -e "${YELLOW}Generating Swagger documentation...${NC}"
swag init -g src/cmd/api/main.go -o docs

echo -e "${YELLOW}Downloading Go dependencies...${NC}"
go mod tidy

echo -e "${YELLOW}Starting services with Docker Compose...${NC}"
$DOCKER_COMPOSE up --build -d

echo ""
echo -e "${GREEN}=== Services Started ===${NC}"
echo -e "API:           http://localhost:9000"
echo -e "Swagger:       http://localhost:9000/swagger/index.html"
echo -e "CockroachDB:   http://localhost:8081 (Admin UI)"
echo ""
echo -e "${YELLOW}To view logs:${NC} $DOCKER_COMPOSE logs -f api"
echo -e "${YELLOW}To stop:${NC} $DOCKER_COMPOSE down"
echo ""
