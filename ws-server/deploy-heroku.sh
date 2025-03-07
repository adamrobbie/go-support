#!/bin/bash

# Heroku deployment script for ws-server

# Check if app name is provided
if [ -z "$1" ]; then
  echo "Usage: ./deploy-heroku.sh <app-name>"
  exit 1
fi

APP_NAME=$1

# Check if Heroku CLI is installed
if ! command -v heroku &> /dev/null; then
  echo "Heroku CLI is not installed. Please install it first."
  echo "Visit: https://devcenter.heroku.com/articles/heroku-cli"
  exit 1
fi

# Check if logged in to Heroku
if ! heroku auth:whoami &> /dev/null; then
  echo "You are not logged in to Heroku. Please login first."
  heroku login
fi

# Check if the app exists
if ! heroku apps:info --app $APP_NAME &> /dev/null; then
  echo "App '$APP_NAME' does not exist. Creating it now..."
  heroku create $APP_NAME
else
  echo "App '$APP_NAME' already exists."
fi

# Set environment variables
echo "Setting environment variables..."
heroku config:set NODE_ENV=production --app $APP_NAME
echo "Do you want to set ALLOWED_ORIGINS? (y/n)"
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
  echo "Enter comma-separated list of allowed origins (e.g., https://example.com,https://www.example.com):"
  read -r origins
  heroku config:set ALLOWED_ORIGINS=$origins --app $APP_NAME
fi

# Build the app
echo "Building the app..."
npm run build

# Deploy to Heroku
echo "Deploying to Heroku..."
git add .
git commit -m "Prepare for Heroku deployment" || true

# Check if we're in a subdirectory
if [ "$(pwd)" != "$(git rev-parse --show-toplevel)" ]; then
  echo "Deploying from subdirectory..."
  git push https://git.heroku.com/$APP_NAME.git `git subtree split --prefix ws-server HEAD`:main --force
else
  echo "Deploying from root directory..."
  git push https://git.heroku.com/$APP_NAME.git main
fi

# Open the app
echo "Deployment completed. Opening the app..."
heroku open --app $APP_NAME

echo "Deployment completed successfully!"
echo "You can monitor your app with: heroku logs --tail --app $APP_NAME" 