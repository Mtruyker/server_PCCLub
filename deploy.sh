#!/bin/bash

# PCClub Server Deployment Script

set -e

echo "🚀 Deploying PCClub Server..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please edit .env file with your production settings before continuing."
    echo "   Especially change JWT_SECRET and ADMIN_API_TOKEN!"
    read -p "Press Enter to continue after editing .env file..."
fi

# Build and start the services
echo "🔨 Building Docker image..."
docker-compose build

echo "🏃 Starting services..."
docker-compose up -d

# Wait for health check
echo "⏳ Waiting for server to be healthy..."
timeout=60
counter=0
while [ $counter -lt $timeout ]; do
    if docker-compose exec -T pcclub-server wget --no-verbose --tries=1 --spider http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ Server is healthy!"
        break
    fi
    sleep 2
    counter=$((counter + 2))
    echo "   Waiting... ($counter/$timeout seconds)"
done

if [ $counter -ge $timeout ]; then
    echo "❌ Server failed to start within $timeout seconds"
    echo "📋 Checking logs..."
    docker-compose logs pcclub-server
    exit 1
fi

echo ""
echo "🎉 PCClub Server deployed successfully!"
echo "📍 Server is running at: http://localhost:8080"
echo "🏥 Health check: http://localhost:8080/health"
echo ""
echo "📋 Useful commands:"
echo "   View logs: docker-compose logs -f pcclub-server"
echo "   Stop server: docker-compose down"
echo "   Restart: docker-compose restart pcclub-server"
echo "   Update: git pull && docker-compose build && docker-compose up -d"