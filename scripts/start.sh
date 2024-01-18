#!/bin/bash

set -e

echo "Start server"
npx nodemon main.go --signal SIGTERM &

echo "Start tailwind watcher"
npx tailwindcss -i ./styles/input.css -o ./public/output.css --watch