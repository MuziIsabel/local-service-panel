#!/usr/bin/env bash
# Start the Agent in development mode.
# Usage: ./scripts/agent-dev.sh

set -e

cd "$(dirname "$0")/../agent"

echo "Starting Local Service Panel Agent..."
echo "Data dir: .data/"
echo "Config: .data/config/agent.json"
echo ""

go run ./cmd/agent -data ../.data
