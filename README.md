# Telegram YouTube Downloader Bot

## Overview

This project houses the source code for a Telegram bot developed in Go. The bot provides the ability to download YouTube videos and audio in various formats, manage user subscriptions through Telegram payments, and check subscription statuses.

## Architecture

The project follows a modular architecture with two main services:

- **Telegram Bot** (`cmd/telegram-bot/`) - Main bot service for handling Telegram interactions
- **User Database API** (`cmd/user-database/`) - REST API service for managing user data and subscriptions

## Key Features

- Download YouTube videos and audio in multiple formats
- Manage user subscriptions and handle payments
- Monitor subscription status and expiry dates
- Performance profiling for CPU and memory usage
- Graceful shutdown capabilities
- REST API for user management
- Automated subscription and traffic management

## Setup Instructions

### Requirements

- Go version 1.21 or higher
- Docker and Docker Compose (for containerized deployment)
- PostgreSQL database
- Telegram Bot Token
- Payment Provider Token (for subscriptions)

### Environment Variables

You can set environment variables in several ways:

#### Option 1: Using .env file (for local development)
Create a `.env` file in the project root:

```env
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_bot_token_here
PROVIDER_TOKEN=your_payment_provider_token

# Database Configuration
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=users
DB_SSLMODE=disable
HOST=localhost
PORT=5432

# API Configuration
DB_URL=https://localhost:8082
```

#### Option 2: Using environment variables directly (for Docker)
Set environment variables in your shell or CI/CD pipeline:

```bash
export TELEGRAM_BOT_TOKEN=your_bot_token_here
export PROVIDER_TOKEN=your_payment_provider_token
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=users
export DB_SSLMODE=disable
export HOST=localhost
export PORT=5432
export DB_URL=https://localhost:8082
```

### Installation Steps

1. Clone this repository
2. Set up your environment variables (see options above)
3. Build and run the services

### Running the Services

#### Using Docker (Recommended)

```bash
# Build and start all services
docker-compose up --build

# Or use Makefile
make docker-up
```

#### Running Locally

```bash
# Build both services
make build

# Run Telegram Bot
make run-telegram-bot

# Run User Database API (in another terminal)
make run-user-database
```

### Bot Commands

- `/start` - Initialize the bot
- `/help` - Display help information
- `/pay` - Purchase a subscription
- `/status` - Check subscription status

### API Endpoints

The User Database API provides the following endpoints:

- `GET /health` - Health check
- `POST /users` - Create a new user
- `GET /users/:username` - Get user information
- `PUT /users/:username` - Update user information
- `DELETE /users/:username` - Delete a user
- `GET /users/:username/exists` - Check if user exists
- `GET /users/:username/subscription` - Get subscription status
- `PUT /users/:username/traffic` - Update user traffic
- `GET /users` - Get all users

## Development

### Building

```bash
# Build both services
make build

# Build individual services
make build-telegram-bot
make build-user-database
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

### Performance Profiling

The bot includes built-in CPU and memory profiling. Profiles are generated when the application receives SIGINT or SIGTERM signals:

- `cpu.prof` - CPU profile
- `mem.prof` - Memory profile

## System Architecture

The bot follows a clean architecture pattern:

- **Presentation Layer**: Telegram bot handlers and HTTP API
- **Business Logic Layer**: Bot logic, downloaders, and schedulers
- **Data Access Layer**: Database repositories and clients
- **Infrastructure Layer**: External services integration

## Contribution Guidelines

We welcome contributions! To contribute:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Special Thanks

Gratitude goes out to the contributors of the libraries utilized in this project, including:
- go-telegram-bot-api
- kkdai/youtube
- joho/godotenv
- gin-gonic/gin
- lib/pq
