# Zero

Zero is an open-source video conferencing application built with Go, designed to provide secure and reliable real-time communication.

## Status

Currently in active development. Features and functionality are subject to change.

## Features

- Peer-to-peer WebRTC video conferencing
- Real-time video streaming with HD support
- Audio capture and monitoring with visual feedback
- Multi-participant session support
- WebSocket-based signaling server
- Camera and microphone controls (pause/resume)
- Live stream statistics and performance metrics
- Visual audio level indicators
- Cross-platform GUI using Fyne
- NAT traversal using STUN servers

## Requirements

- Go 1.25.3 or higher
- Webcam device
- Audio input device (microphone)

## Dependencies

Zero is built using the following key technologies:

- **Fyne v2** - Cross-platform GUI framework
- **Pion WebRTC** - WebRTC implementation for Go
- **Pion MediaDevices** - Media device access and streaming
- **Gorilla WebSocket** - WebSocket library for signaling
- **Google UUID** - Session and peer ID generation

For a complete list of dependencies, see `go.mod`.

## Installation

1. Clone the repository:
```bash
git clone https://github.com/javanhut/zero.git
cd zero
```

2. Install dependencies:
```bash
go mod download
```

3. Start the signaling server (in a separate terminal):
```bash
go run cmd/signaling/main.go
```

4. Run the application:
```bash
go run main.go
```

## Usage

### Prerequisites

Ensure the signaling server is running:
```bash
go run cmd/signaling/main.go
```

The server will start on `localhost:8080` by default.

### Starting a New Session

1. Launch the application
2. Click "Start New Session"
3. A unique session ID will be generated and displayed
4. Share the session ID with participants
5. Your video stream will start automatically
6. WebRTC connections will establish when peers join

### Joining an Existing Session

1. Launch the application
2. Enter the session ID in the text field
3. Click "Connect"
4. Your video stream will start
5. WebRTC connection will be established with existing peers
6. You'll be able to see and hear other participants

### Controls

- **Camera On/Off** - Toggle video streaming
- **Audio On/Off** - Mute/unmute microphone
- **Stats** - View detailed stream statistics including:
  - Stream status (Active/Stopped)
  - Video status (Active/Paused)
  - Audio status (Active/Muted)
  - Resolution
  - Frame rate (FPS)
  - Total frames processed
  - Session duration
  - Current audio level (dB)

### Audio Indicator

The visual audio meter displays real-time microphone input levels:
- Green: Low volume
- Yellow: Medium volume
- Red: High volume

Circle size increases with volume intensity.

## Project Structure

```
Zero/
├── camera/         # Video and audio capture functionality
├── gui/            # User interface implementation
├── sessionmanager/ # Session creation and management
├── signaling/      # WebSocket signaling server and client
├── webrtc/         # WebRTC peer connection management
├── sfu/            # ION SFU integration (stub)
├── cmd/
│   └── signaling/  # Signaling server executable
├── docs/           # Documentation
├── config.yaml     # Configuration file
├── main.go         # Application entry point
├── go.mod          # Go module definition
└── go.sum          # Dependency checksums
```

## Development

### Testing

```bash
go test ./...
```

### Building

Note: Builds are handled via workflow automation. Do not create local release builds unless explicitly instructed.

## Contributing

Contributions are welcome. Please ensure:

- Code follows Go best practices
- All tests pass before submitting
- Documentation is updated for new features

## License

To be determined

## Roadmap

- [x] WebRTC peer-to-peer connections
- [x] Multi-participant support
- [x] WebSocket signaling server
- [x] STUN server integration
- [ ] ION SFU integration for scalability
- [ ] Remote video display in GUI
- [ ] Screen sharing
- [ ] Chat functionality
- [ ] Recording capabilities
- [ ] Enhanced security (TLS/WSS, authentication)
- [ ] TURN server support for better NAT traversal
- [ ] Simulcast and bandwidth adaptation

## Support

For issues, questions, or contributions, please visit the GitHub repository.

## Architecture

For detailed technical architecture documentation, see:

- [WebRTC Architecture](docs/WEBRTC_ARCHITECTURE.md)
- [Signaling Protocol](docs/SIGNALING_PROTOCOL.md)

### Key Components

1. **Signaling Server**: WebSocket server for peer coordination
2. **WebRTC Manager**: Manages peer-to-peer connections
3. **Session Manager**: Tracks active sessions and participants
4. **Media Capture**: Camera and microphone access via Pion MediaDevices
5. **GUI**: Cross-platform interface built with Fyne

## Network Requirements

### Firewall

- Outbound: WebSocket connection to signaling server (port 8080)
- Outbound: STUN server access (UDP port 19302)
- Outbound/Inbound: WebRTC media (UDP dynamic ports)

### NAT Traversal

Zero uses Google's public STUN servers for NAT traversal. For networks with symmetric NAT or strict firewalls, you may need to configure TURN servers.

## Troubleshooting

### Signaling Server Connection Failed

- Ensure signaling server is running: `go run cmd/signaling/main.go`
- Check that port 8080 is not in use by another application
- Verify firewall allows outbound connections to localhost:8080

### WebRTC Connection Issues

- Check that UDP traffic is allowed through firewall
- Verify STUN servers are accessible
- For restrictive networks, configure TURN servers in `config.yaml`

### No Video or Audio

- Check camera and microphone permissions
- Verify devices are not in use by another application
- Check system audio/video settings

For more detailed troubleshooting, see [docs/WEBRTC_ARCHITECTURE.md](docs/WEBRTC_ARCHITECTURE.md).

## Acknowledgments

Built with the Pion WebRTC stack and Fyne GUI framework.
