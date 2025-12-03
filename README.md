# Real-time Chat Server

A high-performance real-time chat server built with Go, demonstrating advanced concurrency patterns with WebSocket, Goroutines, and Channels.

## Features

✅ **Real-time Messaging** - Instant message delivery using WebSocket  
✅ **Broadcast Messages** - Send messages to all connected users  
✅ **Private Messaging** - Direct messages between users (infrastructure ready)  
✅ **Message History** - Last 100 messages stored in Redis Lists  
✅ **Multi-Instance Support** - Horizontal scaling via Redis Pub/Sub  
✅ **User Presence** - Real-time online tracking with Redis Sets & TTL  
✅ **Rate Limiting** - Prevent spam (10 messages/10 seconds per user)  
✅ **Pub/Sub Pattern** - Event-driven architecture with Go Channels  
✅ **Concurrent Processing** - Each client runs in separate Goroutines  
✅ **Graceful Disconnection** - Proper cleanup on client disconnect  
✅ **Modern UI** - Responsive web interface with gradient design

## Tech Stack

- **Language**: Go 1.25+
- **WebSocket**: Gorilla WebSocket
- **Concurrency**: Goroutines & Channels
- **Cache/Store**: Redis 7
- **Frontend**: Vanilla HTML/CSS/JavaScript
- **Containerization**: Docker

## Architecture Highlights

### Goroutines & Channels in Action

This project demonstrates Go's concurrency model:

1. **Hub Goroutine** - Central message broker running in `select` loop
   - Handles client registration/unregistration
   - Broadcasts messages to all clients
   - Routes private messages
   - Subscribes to Redis Pub/Sub for multi-instance support

2. **Client Goroutines** - Each client spawns 2 Goroutines:
   - **ReadPump**: Reads messages from WebSocket → sends to Hub
   - **WritePump**: Reads from channel → writes to WebSocket

3. **Presence Heartbeat Goroutine** - Updates Redis presence every 10s

4. **Channels** - Non-blocking communication:
   - `register chan` - New client connections
   - `unregister chan` - Client disconnections  
   - `broadcast chan` - Message distribution
   - `send chan` - Per-client message queue (buffered 256)
   - Redis Pub/Sub channel - Cross-instance messaging

### Redis Integration

**Message History** (Redis Lists)
- Store last 100 messages with LPUSH/LTRIM
- New clients receive history on connect
- 24-hour TTL for automatic cleanup

**Horizontal Scaling** (Redis Pub/Sub)
- Multiple server instances share messages
- Each instance broadcasts to Redis
- All instances receive and forward to local clients

**User Presence** (Redis Keys with TTL)
- Track online users across all instances
- 30-second TTL with heartbeat keepalive
- Auto-removal on disconnect or timeout

**Rate Limiting** (Redis Counters)
- Limit: 10 messages per 10 seconds per user
- Atomic increment with INCR
- Prevents spam and abuse

## Project Structure

```
.
├── cmd/server/          # Application entry point
│   └── main.go         # HTTP server & WebSocket upgrade
├── internal/
│   ├── hub/            # Pub/Sub hub with channels
│   │   ├── hub.go      # Central message broker (Goroutine)
│   │   └── helpers.go  # History, presence, heartbeat
│   ├── client/         # WebSocket client handler
│   │   └── client.go   # Read/Write pumps (2 Goroutines per client)
│   └── redis/          # Redis integrations
│       ├── client.go          # Redis connection
│       ├── message_history.go # Message storage (Lists)
│       ├── pubsub.go          # Multi-instance Pub/Sub
│       ├── presence.go        # Online users (Keys+TTL)
│       └── rate_limiter.go    # Spam prevention (Counters)
├── pkg/
│   └── models/         # Domain models
│       └── message.go  # Message types & User struct
└── public/             # Frontend assets
    ├── index.html      # Chat UI
    ├── style.css       # Styling
    └── script.js       # WebSocket client
```

## Getting Started

### Prerequisites

- Go 1.25+ installed
- Redis 7+ (or use Docker)
- Docker & Docker Compose (optional)

### Option 1: Run with Docker Compose (Recommended)

```bash
# Build and run (includes Redis)
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Option 2: Run with Go

1. Clone the repository
2. Start Redis:
   ```bash
   docker run -d -p 6379:6379 redis:7-alpine
   # Or use local Redis installation
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Run the server:
   ```bash
   export REDIS_ADDR=localhost:6379
   go run cmd/server/main.go
   ```

### Option 3: Manual with Docker Compose

```bash
# Start only Redis
docker-compose up -d redis

# Run server locally
go run cmd/server/main.go
```

## Environment Variables

```bash
SERVER_PORT=8080              # HTTP server port
REDIS_ADDR=localhost:6379     # Redis address
REDIS_PASSWORD=               # Redis password (if any)
```

## Testing Multi-Instance Setup

To test horizontal scaling with Redis Pub/Sub:

```bash
# Terminal 1 - Start Redis
docker-compose up -d redis

# Terminal 2 - Server instance 1 (port 8080)
export SERVER_PORT=8080
export REDIS_ADDR=localhost:6379
go run cmd/server/main.go

# Terminal 3 - Server instance 2 (port 8081)
export SERVER_PORT=8081
export REDIS_ADDR=localhost:6379
go run cmd/server/main.go

# Terminal 4 - Test
# Open http://localhost:8080 and http://localhost:8081
# Connect users to different ports
# Messages sent to one port appear on the other!
```
   go run cmd/server/main.go
   ```
4. Open browser at `http://localhost:8080`

### Option 2: Run with Docker

```bash
# Build and run
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

## Usage

1. Open `http://localhost:8080` in your browser
2. Enter a username and click "Join Chat"
3. Start sending messages!
4. Open multiple browser tabs/windows to simulate multiple users

### Access from Other Devices (Phone, Tablet, etc.)

The chat server supports multi-device access over the same local network. This is perfect for testing or using the chat across your devices at home or office.

#### Step-by-Step Guide:

**1. Find Your Server's IP Address**

On your server/laptop, run:
```bash
ip addr show | grep "inet " | grep -v "127.0.0.1"
# Example output: 192.168.1.100
```

Or use the quick command:
```bash
hostname -I | awk '{print $1}'
```

**2. Ensure Both Devices Are Connected**
- Server (laptop) and client device (phone/tablet) must be on the **same WiFi network**
- Make sure WiFi isolation is disabled (usually enabled on guest networks)

**3. Open Firewall Port (Security)**

The server needs port 8080 open for external connections. We provide a helper script for easy management:

```bash
# Open port temporarily (for testing)
./firewall-control.sh open

# Check port status
./firewall-control.sh status

# Close port after testing (recommended for security)
./firewall-control.sh close
```

See [Firewall Management](#firewall-management) section below for detailed usage.

**4. Access from Your Device**

Open a browser on your phone/tablet and navigate to:
```
http://YOUR_SERVER_IP:8080
```

For example, if your server IP is `192.168.1.100`:
```
http://192.168.1.100:8080
```

**5. Start Chatting!**
- Enter your username on the phone
- Open the laptop browser at `http://localhost:8080`
- Chat between devices in real-time!

#### Features Across Devices:
- ✅ **Real-time sync** - Messages appear instantly on all devices
- ✅ **Message history** - New devices see last 50 messages
- ✅ **Online presence** - See who's online across all instances
- ✅ **Mobile responsive** - UI optimized for phone screens
- ✅ **WebSocket auto-detection** - No config needed for different IPs

#### Troubleshooting:

**"Site can't be reached" on phone:**
1. Check if port is open: `./firewall-control.sh status`
2. Verify both devices on same WiFi
3. Try pinging: `ping YOUR_SERVER_IP` from phone
4. Check if server is running: `curl http://localhost:8080`

**Connection refused:**
- Firewall is likely blocking - run `./firewall-control.sh open`
- Make sure server is running on `0.0.0.0:8080` (not `127.0.0.1`)

**Messages not syncing:**
- Check Redis is running: `docker ps | grep redis`
- Check logs: `tail -f /tmp/chat-server.log`

## Firewall Management

We provide a convenient script to manage firewall rules for the chat server port (8080).

### Script Usage

The `firewall-control.sh` script supports the following commands:

```bash
# Open port temporarily (for testing only)
# Port closes automatically after system reboot
./firewall-control.sh open

# Close port (recommended after testing)
./firewall-control.sh close

# Check if port is open or closed
./firewall-control.sh status

# Open port permanently (with confirmation prompt)
# Use only if you want the server always accessible
./firewall-control.sh permanent-open

# Close port permanently
./firewall-control.sh permanent-close
```

### Security Recommendations

🔒 **Best Practice:**
1. Use `open` for testing sessions
2. Always `close` when done
3. Avoid `permanent-open` unless necessary

**Why?** Opening port 8080 permanently exposes your chat server to your entire local network. Only do this if:
- You trust all devices on your network
- You're running on a private/home network
- You want the server always accessible

**Example Workflow:**
```bash
# Start testing
./firewall-control.sh open

# Test on phone/tablet...

# Done testing
./firewall-control.sh close
```

### Manual Firewall Commands

If you prefer manual control or the script doesn't work:

**For firewalld (Fedora, RHEL, CentOS):**
```bash
# Temporary (until reboot)
sudo firewall-cmd --add-port=8080/tcp
sudo firewall-cmd --remove-port=8080/tcp

# Permanent
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
sudo firewall-cmd --remove-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

**For ufw (Ubuntu, Debian):**
```bash
sudo ufw allow 8080/tcp
sudo ufw delete allow 8080/tcp
sudo ufw status
```

**For iptables:**
```bash
sudo iptables -I INPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables -D INPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables-save
```

## How It Works

### Connection Flow

```
Client connects → WebSocket upgrade → Create Client instance
    ↓
Register with Hub (via channel) → Hub adds to clients map
    ↓
Spawn 2 Goroutines:
  - ReadPump: Listen for incoming messages
  - WritePump: Send outgoing messages
    ↓
Hub runs in select loop, distributing messages via channels
```

### Message Flow

```
User types message → WebSocket send → ReadPump receives
    ↓
ReadPump → Hub.Broadcast (channel) → Hub select loop
    ↓
Hub iterates clients → Send to each client's send channel
    ↓
WritePump reads from send channel → WebSocket write to client
```

## Key Concurrency Patterns

### 1. Fan-out (Broadcast)
One message from Hub → Multiple clients simultaneously

### 2. Non-blocking Send
```go
select {
case c.send <- message:
default:
    close(c.send) // Prevent blocking
}
```

### 3. Graceful Shutdown
```go
defer func() {
    hub.Unregister(client)
    conn.Close()
}()
```

## Testing

```bash
# Build the application
go build -o bin/chat-server ./cmd/server

# Run tests (if added)
go test ./...
```

## Future Enhancements

- [ ] Message persistence with database
- [ ] Multiple chat rooms
- [ ] User authentication with JWT
- [ ] File/image sharing
- [ ] Typing indicators
- [ ] Message history

## License

MIT
