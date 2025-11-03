# Zero WebRTC Architecture

## Overview

Zero uses a peer-to-peer WebRTC architecture with WebSocket signaling to enable multi-participant video conferencing. This document describes the technical architecture and implementation details.

## Architecture Components

### 1. Signaling Layer (`signaling/`)

The signaling layer handles the coordination between peers before WebRTC connections are established.

#### Components

- **Server** (`signaling/server.go`): WebSocket server that coordinates peer connections
  - Manages sessions and connected clients
  - Routes signaling messages between peers
  - Notifies peers when others join or leave

- **Client** (`signaling/client.go`): WebSocket client for peer communication
  - Connects to signaling server
  - Sends/receives SDP offers, answers, and ICE candidates
  - Event-based message handling

- **Messages** (`signaling/messages.go`): Protocol definitions
  - Join/Leave messages
  - SDP Offer/Answer
  - ICE Candidate exchange
  - Peer joined/left notifications

#### Message Flow

```
Client A                Signaling Server           Client B
   |                           |                        |
   |----Join(SessionID)------->|                        |
   |                           |<----Join(SessionID)----|
   |<--PeerJoined(B)-----------|                        |
   |                           |----PeerJoined(A)------>|
   |----Offer(SDP)------------>|                        |
   |                           |----Offer(SDP)--------->|
   |                           |<---Answer(SDP)---------|
   |<--Answer(SDP)-------------|                        |
   |----ICE Candidate--------->|                        |
   |                           |----ICE Candidate------>|
```

### 2. WebRTC Layer (`webrtc/`)

Manages peer-to-peer WebRTC connections and media streaming.

#### Components

- **Config** (`webrtc/config.go`): WebRTC configuration
  - ICE server configuration (STUN/TURN)
  - Default configuration provider

- **Peer** (`webrtc/peer.go`): Individual peer connection
  - Wraps `pion/webrtc` PeerConnection
  - Handles ICE candidates
  - Manages local and remote tracks
  - Connection state monitoring

- **Manager** (`webrtc/manager.go`): Multi-peer connection manager
  - Creates and manages multiple peer connections
  - Integrates with signaling layer
  - Handles offer/answer negotiation
  - Distributes local tracks to all peers

#### Connection Establishment

1. Peer A joins session via signaling server
2. Peer B joins same session
3. Signaling server notifies Peer A of Peer B
4. Manager creates PeerConnection for Peer B
5. Peer A creates and sends SDP offer
6. Peer B receives offer, creates answer
7. ICE candidates exchanged
8. Media flows directly peer-to-peer

### 3. Session Management (`sessionmanager/`)

Manages conference sessions and participants.

#### Features

- Create new sessions with unique IDs
- Join existing sessions
- Track peers in each session
- Thread-safe session operations
- Peer connection state tracking

#### Session Lifecycle

```
Create Session -> Peers Join -> Peers Connect -> Peers Leave -> Session Ends
```

### 4. Media Capture (`camera/`)

Handles local camera and microphone access.

#### Features

- Video capture with configurable resolution (HD, FullHD)
- Audio capture with level monitoring
- Pause/resume video and audio
- Stream statistics (FPS, frame count, duration)
- WebRTC track creation

#### Media Flow

```
Device -> MediaDevices API -> VideoStream -> GUI Display
                                    |
                                    v
                              WebRTC Tracks -> Remote Peers
```

### 5. SFU Integration (`sfu/`)

Placeholder for ION SFU integration (future enhancement).

#### Current State

- Stub implementation
- Ready for ION SFU SDK integration
- Will enable scalable multi-party calls (3+ participants)

#### SFU vs Peer-to-Peer

**Peer-to-Peer (Current)**:
- Direct connections between peers
- Best for 2-3 participants
- Higher bandwidth for each peer

**SFU (Future)**:
- Central media router
- Better for 3+ participants
- Lower client bandwidth
- Server infrastructure required

## Network Architecture

### Topology

```
                    Signaling Server
                    (WebSocket :8080)
                           |
          +----------------+----------------+
          |                                 |
     Client A                          Client B
    (WebRTC Peer)                    (WebRTC Peer)
          |                                 |
          +-------WebRTC Connection---------+
                  (Peer-to-Peer)
```

### Ports and Protocols

- **Signaling**: WebSocket on port 8080 (configurable)
- **WebRTC Media**: UDP (dynamic ports, negotiated via ICE)
- **STUN**: UDP 19302 (Google STUN servers)
- **TURN**: Not configured (optional for NAT traversal)

## Security Considerations

### Current Implementation

- WebSocket connections (ws://) - not encrypted
- No authentication on signaling server
- Sessions identified by UUID only

### Recommended Enhancements

1. **TLS/WSS**: Use secure WebSocket (wss://)
2. **Authentication**: Add user authentication
3. **Session Passwords**: Optional password protection for sessions
4. **DTLS-SRTP**: Already enabled by WebRTC for media encryption
5. **Rate Limiting**: Prevent abuse of signaling server

## Scalability

### Current Limits

- Peer-to-peer: Works well for 2-3 participants
- Each peer maintains N-1 connections (N = total peers)
- Bandwidth scales linearly per peer

### Future Scalability (with SFU)

- SFU can handle 10-100+ participants
- Each client maintains 1 connection to SFU
- Server bandwidth scales linearly
- Client bandwidth remains constant

## Configuration

Configuration file: `config.yaml`

```yaml
signaling:
  server_address: "localhost:8080"
  
webrtc:
  ice_servers:
    - urls: "stun:stun.l.google.com:19302"

sfu:
  enabled: false
  
media:
  video:
    resolution: "HD"
  audio:
    sample_rate: 48000
```

## Testing Scenarios

### Basic Connection Test

1. Start signaling server: `go run cmd/signaling/main.go`
2. Start Client A: `go run main.go`
3. Create new session in Client A
4. Start Client B: `go run main.go`
5. Join session from Client B using session ID
6. Verify video/audio connection

### Multi-Peer Test

1. Start signaling server
2. Create session with Client A
3. Join with Client B
4. Join with Client C
5. Verify all peers can see each other

## Troubleshooting

### Common Issues

1. **Cannot connect to signaling server**
   - Ensure server is running on correct port
   - Check firewall settings
   - Verify WebSocket URL in config

2. **ICE connection fails**
   - Check STUN server accessibility
   - May need TURN server for restrictive NATs
   - Verify UDP traffic is allowed

3. **No video/audio**
   - Check camera/microphone permissions
   - Verify tracks are added to peer connection
   - Check browser console for errors

## Future Enhancements

1. **ION SFU Integration**: Complete SFU client implementation
2. **Simulcast**: Multiple quality levels
3. **Screen Sharing**: Desktop capture
4. **Recording**: Server-side recording capability
5. **Chat**: Text messaging during calls
6. **Bandwidth Adaptation**: Dynamic quality adjustment
7. **E2E Encryption**: Optional end-to-end encryption
8. **Mobile Support**: iOS and Android clients
