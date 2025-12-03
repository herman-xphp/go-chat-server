# Real-time Chat Server

A high-performance real-time chat server built with Go, WebSocket, Goroutines, and Channels.

## Features

- Real-time messaging with WebSocket
- Chat rooms (broadcast messaging)
- Private messaging (direct messages)
- Online/offline user status
- Pub/Sub pattern with Go Channels
- Concurrent message handling with Goroutines

## Tech Stack

- **Language**: Go 1.21+
- **WebSocket**: Gorilla WebSocket
- **Concurrency**: Goroutines & Channels
- **Frontend**: Vanilla HTML/CSS/JavaScript

## Project Structure

```
.
├── cmd/server/          # Application entry point
├── internal/
│   ├── hub/            # Pub/Sub hub with channels
│   └── client/         # WebSocket client handler
├── pkg/
│   └── models/         # Domain models
└── public/             # Frontend assets
```

## Getting Started

### Prerequisites

- Go 1.21+ installed

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Run the server:
   ```bash
   go run cmd/server/main.go
   ```
4. Open browser at `http://localhost:8080`

## License

MIT
