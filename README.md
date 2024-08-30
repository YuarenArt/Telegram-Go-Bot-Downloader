## Overview

This repository contains the source code for a Telegram bot written in Go. The bot provides functionality for downloading YouTube videos and audio in various formats, handling payments for premium features, and managing user subscriptions.

## Features

- Download YouTube videos and audio in multiple formats.
- Manage user subscriptions through Telegram payments.
- Check subscription status and expiration dates.
- Profile the bot's performance using CPU and memory profiling.
- Graceful shutdown handling.

## Getting Started

### Prerequisites

- Go version 1.21+
- Docker (for running the bot in a container)

### Installation

1. Clone this repository.
2. Set up your environment variables in `.env` for Telegram bot token, database URL, provider token, etc.
3. Build and run the bot using Docker.

### Usage

#### Running the Bot

```sh
docker-compose up
Testing
To test the bot's performance, you can use the profiling features built into the bot. By default, the bot will generate a CPU profile (cpu.prof) and a memory profile (mem.prof) when it receives a SIGINT or SIGTERM signal.

Commands
/start: Start the bot.
/help: Display help information.
/pay: Pay for a subscription.
/status: Check subscription status.
Architecture
The bot is structured using the Model-View-Controller (MVC) pattern. It communicates with the Telegram API using the go-telegram-bot-api library and interacts with the database via a custom client.

Contributing
Contributions are welcome! Please follow these guidelines:

Fork the repository.
Create a new branch for your feature or bug fix.
Make your changes and commit them.
Push to your fork and submit a pull request.
License
This project is licensed under the MIT License - see the LICENSE file for details.

Acknowledgments
Thanks to the contributors of the libraries used in this project, including go-telegram-bot-api, kkdai/youtube, and joho/godotenv.
