#!/bin/bash
set -e

echo "🚀 Starting deployment..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${RED}❌ Error: .env file not found!${NC}"
    echo "Please create .env file with required variables"
    exit 1
fi

# Load environment variables
source .env

echo -e "${YELLOW}📦 Building Docker images...${NC}"
docker compose -f docker-compose.prod.yml build --no-cache

echo -e "${YELLOW}🔄 Stopping old containers...${NC}"
docker compose -f docker-compose.prod.yml down

echo -e "${YELLOW}🚀 Starting new containers...${NC}"
docker compose -f docker-compose.prod.yml up -d

echo -e "${YELLOW}⏳ Waiting for services to be healthy...${NC}"
sleep 10

# Check if containers are running
if docker compose -f docker-compose.prod.yml ps | grep -q "Up"; then
    echo -e "${GREEN}✅ Deployment successful!${NC}"

    # Show running containers
    echo -e "\n${YELLOW}📊 Running containers:${NC}"
    docker compose -f docker-compose.prod.yml ps

    # Show logs
    echo -e "\n${YELLOW}📝 Recent logs:${NC}"
    docker compose -f docker-compose.prod.yml logs --tail=50
else
    echo -e "${RED}❌ Deployment failed! Containers are not running.${NC}"
    docker compose -f docker-compose.prod.yml logs
    exit 1
fi

# Connect to Traefik network if needed
if docker network ls | grep -q traefik_default; then
    echo -e "${YELLOW}🔗 Connecting to Traefik network...${NC}"
    docker network connect traefik_default alias-backend-1 2>/dev/null || echo "Backend already connected"
    docker network connect traefik_default alias-frontend-1 2>/dev/null || echo "Frontend already connected"

    # Restart Traefik to pick up new services
    docker restart traefik 2>/dev/null || echo "Traefik not found or not running"
fi

echo -e "${GREEN}🎉 Deployment complete!${NC}"
