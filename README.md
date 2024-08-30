## Overview

This project houses the source code for a Telegram bot developed in Go. The bot provides the ability to download YouTube videos and audio in various formats, manage user subscriptions through Telegram payments, and check subscription statuses.

## Key Features

- Download YouTube videos and audio in multiple formats.
- Manage user subscriptions and handle payments.
- Monitor subscription status and expiry dates.
- Performance profiling for CPU and memory usage.
- Graceful shutdown capabilities.

## Setup Instructions

### Requirements

- Go version 1.21 or higher.
- Docker (for containerized deployment).

### Installation Steps

1. Clone this repository.
2. Set up your environment variables in the `.env` file for the Telegram bot token, database URL, provider token, etc.
3. Build and run the bot using Docker.

### Running the Bot

docker-compose up

To evaluate the bot's performance, you can leverage the integrated profiling features. By default, the bot generates a CPU profile (cpu.prof) and a memory profile (mem.prof) upon receiving a SIGINT or SIGTERM signal.

### Bot Commands
/start: Initiate the bot.
/help: Display help information.
/pay: Purchase a subscription.
/status: Check subscription status.
System Architecture
The bot follows the Model-View-Controller (MVC) design pattern. It interacts with the Telegram API through the go-telegram-bot-api library and communicates with the database using a custom client.

### Contribution Guidelines
We welcome contributions! To contribute:

### Special Thanks
Gratitude goes out to the contributors of the libraries utilized in this project, including go-telegram-bot-api, kkdai/youtube, and joho/godotenv.
