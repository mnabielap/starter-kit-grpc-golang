#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

echo "Starting gRPC Application..."

# Execute the CMD passed from Dockerfile or docker run
exec "$@"