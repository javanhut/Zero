# Zero - Quick Start Guide

This guide will help you get Zero up and running quickly for local testing.

## Prerequisites

- Go 1.25.3 or higher
- Webcam device
- Microphone
- Linux/macOS/Windows operating system

## Quick Setup (Local Testing)

### Step 1: Install Dependencies

```bash
cd /path/to/Zero
go mod download
```

### Step 2: Start the Signaling Server

Open a terminal window and run:

```bash
go run cmd/signaling/main.go
```

You should see:
```
2024/XX/XX XX:XX:XX Starting Zero signaling server on :8080
2024/XX/XX XX:XX:XX Signaling server starting on :8080
```

Leave this terminal running.

### Step 3: Start First Client (Alice)

Open a new terminal window and run:

```bash
go run main.go
```

1. Click "Start New Session"
2. Copy the session ID displayed in the text field
3. Your camera should activate

### Step 4: Start Second Client (Bob)

Open another terminal window and run:

```bash
go run main.go
```

1. Paste the session ID from Alice
2. Click "Connect"
3. Your camera should activate
4. WebRTC connection will be established

## Testing the Connection

### Verify Signaling

In the signaling server terminal, you should see:
```
Created new session: <session-id>
Added client <peer-id-alice> to session <session-id>
Client <peer-id-alice> joined session <session-id>
Added client <peer-id-bob> to session <session-id>
Client <peer-id-bob> joined session <session-id>
```

### Verify WebRTC

In each client terminal, you should see:
```
Connected to signaling server: ws://localhost:8080/ws
Peer joined: User_XXXXXXXX (<peer-id>)
Created peer connection for: <peer-id>
```

### Test Controls

- **Camera On/Off**: Toggle video
- **Audio On/Off**: Toggle microphone
- **Stats**: View stream statistics

## Multi-Peer Testing (3+ Participants)

1. Keep signaling server running
2. Start additional clients with "Connect"
3. Use the same session ID
4. Each client will establish connections with all others

## Troubleshooting

### Signaling Server Won't Start

**Error**: `bind: address already in use`

**Solution**: Another process is using port 8080.

```bash
lsof -i :8080
kill -9 <PID>
```

Or change the port in `cmd/signaling/main.go`:
```go
addr := ":8081"  // Change from :8080
```

### Client Can't Connect to Signaling Server

**Error**: `Failed to connect to signaling server`

**Solution**: 
1. Verify server is running
2. Check server URL in `gui/gui.go`:
```go
signalingServerURL := "ws://localhost:8080/ws"
```

### Camera Not Starting

**Error**: `Failed to start camera`

**Possible Causes**:
- Camera in use by another application
- No camera permission
- Camera not detected

**Solutions**:
- Close other applications using camera
- Grant camera permissions
- Check `dmesg` or system logs for device errors

### No WebRTC Connection

**Symptoms**: Clients connect to signaling but no peer connection

**Debug Steps**:

1. Check client logs for:
```
Created peer connection for: <peer-id>
ICE connection state: checking
ICE connection state: connected
```

2. Verify STUN server access:
```bash
nc -u -v stun.l.google.com 19302
```

3. Check firewall allows UDP traffic

4. Enable verbose WebRTC logs (if needed)

### Audio/Video Quality Issues

**Solutions**:
- Close unnecessary applications
- Check CPU usage
- Adjust resolution in camera.go:
```go
Resolution = map[string]ScreenSize{
    "HD":     {1280, 720},   // Try lower resolution
    "SD":     {640, 480},     // Better for slow networks
}
```

## Configuration

### Change Signaling Server Port

Edit `cmd/signaling/main.go`:
```go
addr := ":8080"  // Change to your port
```

Edit `gui/gui.go`:
```go
signalingServerURL := "ws://localhost:8080/ws"  // Update URL
```

### Change STUN Servers

Edit `webrtc/config.go`:
```go
ICEServers: []webrtc.ICEServer{
    {
        URLs: []string{"stun:stun.l.google.com:19302"},
    },
    {
        URLs: []string{"stun:your-stun-server.com:3478"},
    },
},
```

### Change Video Resolution

Edit `gui/gui.go`:
```go
stream, err := camera.StartVideoStream("HD", updateVideo)
// Options: "HD" (1280x720), "FullHD" (1920x1080)
```

## Network Setup

### Testing Over LAN

1. Find server IP:
```bash
ip addr show  # Linux
ipconfig      # Windows
ifconfig      # macOS
```

2. Update client signaling URL:
```go
signalingServerURL := "ws://192.168.1.100:8080/ws"
```

3. Ensure firewall allows:
   - TCP 8080 (signaling)
   - UDP (WebRTC media)

### Testing Over Internet

Requirements:
- Public IP or domain name
- Port forwarding configured
- Optional: TURN server for NAT traversal

See [WEBRTC_ARCHITECTURE.md](WEBRTC_ARCHITECTURE.md) for detailed setup.

## Building for Distribution

### Linux

```bash
go build -o zero main.go
go build -o zero-signaling cmd/signaling/main.go
```

### Windows

```bash
GOOS=windows GOARCH=amd64 go build -o zero.exe main.go
GOOS=windows GOARCH=amd64 go build -o zero-signaling.exe cmd/signaling/main.go
```

### macOS

```bash
GOOS=darwin GOARCH=amd64 go build -o zero-macos main.go
GOOS=darwin GOARCH=amd64 go build -o zero-signaling-macos cmd/signaling/main.go
```

## Next Steps

- Read [WEBRTC_ARCHITECTURE.md](WEBRTC_ARCHITECTURE.md) for technical details
- Review [SIGNALING_PROTOCOL.md](SIGNALING_PROTOCOL.md) for protocol specification
- Check the main [README.md](../README.md) for feature roadmap

## Getting Help

If you encounter issues:

1. Check logs in terminal output
2. Review troubleshooting section above
3. Consult documentation in `docs/` directory
4. Open an issue on GitHub

## Performance Tips

- Close unnecessary browser tabs and applications
- Use wired ethernet connection when possible
- Test with 2 peers first, then scale up
- Monitor CPU and network usage
- Consider using SFU for 4+ participants (future feature)

Happy conferencing!
