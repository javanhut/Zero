# Zero

Zero is an open-source video conferencing application built with Go, designed to provide secure and reliable real-time communication.

## Status

Currently in active development. Features and functionality are subject to change.

## Features

- Real-time video streaming with HD support
- Audio capture and monitoring with visual feedback
- Session-based communication system
- Camera and microphone controls (pause/resume)
- Live stream statistics and performance metrics
- Visual audio level indicators
- Cross-platform GUI using Fyne

## Requirements

- Go 1.25.3 or higher
- Webcam device
- Audio input device (microphone)

## Dependencies

Zero is built using the following key technologies:

- **Fyne v2** - Cross-platform GUI framework
- **Pion WebRTC** - WebRTC implementation for Go
- **Pion MediaDevices** - Media device access and streaming
- **Google UUID** - Session ID generation

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

3. Run the application:
```bash
go run main.go
```

## Usage

### Starting a New Session

1. Launch the application
2. Click "Start New Session"
3. A unique session ID will be generated
4. Share the session ID with participants
5. Your video stream will start automatically

### Joining an Existing Session

1. Launch the application
2. Enter the session ID in the text field
3. Click "Connect"
4. Your video stream will start upon successful connection

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

- WebRTC peer-to-peer connections
- Multi-participant support
- Screen sharing
- Chat functionality
- Recording capabilities
- Enhanced security features
- TURN/STUN server integration

## Support

For issues, questions, or contributions, please visit the GitHub repository.

## Acknowledgments

Built with the Pion WebRTC stack and Fyne GUI framework.
