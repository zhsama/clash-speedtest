#!/bin/bash

# Start all services for Clash SpeedTest development

echo "🚀 Starting Clash SpeedTest Development Environment..."
echo ""
echo "📦 Frontend will be available at: http://localhost:4321"
echo "🔧 Backend API will be available at: http://localhost:8090"
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Run turbo dev with better output
exec npm run dev