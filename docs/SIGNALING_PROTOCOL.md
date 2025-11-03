# Zero Signaling Protocol

## Overview

The Zero signaling protocol uses WebSocket connections with JSON messages to coordinate WebRTC peer connections.

## Connection

- **Protocol**: WebSocket
- **Default URL**: `ws://localhost:8080/ws`
- **Transport**: TCP

## Message Format

All messages follow this JSON structure:

```json
{
  "type": "message_type",
  "session_id": "uuid",
  "peer_id": "uuid",
  "username": "User_12345678",
  "payload": {}
}
```

### Fields

- `type` (string, required): Message type identifier
- `session_id` (string, required): Session/room identifier
- `peer_id` (string, required): Sender's unique peer ID
- `username` (string, optional): Display name
- `payload` (object, optional): Message-specific data

## Message Types

### 1. Join

Sent when a peer joins a session.

**Direction**: Client -> Server

```json
{
  "type": "join",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "peer_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "username": "User_7c9e6679",
  "payload": {
    "username": "User_7c9e6679"
  }
}
```

**Server Action**:
- Add peer to session
- Broadcast `peer_joined` to existing peers

### 2. Leave

Sent when a peer leaves a session.

**Direction**: Client -> Server

```json
{
  "type": "leave",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "peer_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"
}
```

**Server Action**:
- Remove peer from session
- Broadcast `peer_left` to remaining peers

### 3. Peer Joined

Broadcast when a new peer joins the session.

**Direction**: Server -> Clients

```json
{
  "type": "peer_joined",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "peer_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "username": "User_7c9e6679",
  "payload": {
    "peer_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "username": "User_7c9e6679"
  }
}
```

**Client Action**:
- Create PeerConnection for new peer
- Send SDP offer

### 4. Peer Left

Broadcast when a peer leaves the session.

**Direction**: Server -> Clients

```json
{
  "type": "peer_left",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "peer_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "payload": {
    "peer_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"
  }
}
```

**Client Action**:
- Close PeerConnection for departed peer
- Remove peer from UI

### 5. Offer

WebRTC SDP offer for connection negotiation.

**Direction**: Client -> Server -> Other Client

```json
{
  "type": "offer",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "peer_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "payload": {
    "sdp": {
      "type": "offer",
      "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n..."
    }
  }
}
```

**Server Action**:
- Forward to all other peers in session

**Recipient Action**:
- Set remote description
- Create SDP answer

### 6. Answer

WebRTC SDP answer in response to offer.

**Direction**: Client -> Server -> Other Client

```json
{
  "type": "answer",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "peer_id": "a1b2c3d4-5678-90ab-cdef-1234567890ab",
  "payload": {
    "sdp": {
      "type": "answer",
      "sdp": "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n..."
    }
  }
}
```

**Server Action**:
- Forward to all other peers in session

**Recipient Action**:
- Set remote description

### 7. Candidate

ICE candidate for connection establishment.

**Direction**: Client -> Server -> Other Client

```json
{
  "type": "candidate",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "peer_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "payload": {
    "candidate": {
      "candidate": "candidate:1 1 udp 2130706431 192.168.1.100 54321 typ host",
      "sdpMLineIndex": 0,
      "sdpMid": "0"
    }
  }
}
```

**Server Action**:
- Forward to all other peers in session

**Recipient Action**:
- Add ICE candidate to peer connection

### 8. Error

Error notification from server.

**Direction**: Server -> Client

```json
{
  "type": "error",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "peer_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "payload": {
    "message": "Invalid session ID"
  }
}
```

## Connection Flow

### New Session Creation

```
Client                          Server
  |                               |
  |-----WebSocket Connect-------->|
  |<----Connection Established----|
  |                               |
  |-----Join Message------------->|
  |                               |
  |<----Acknowledged--------------|
```

### Peer Joining Existing Session

```
Client A        Server          Client B
  |               |                |
  |               |<--Join---------|
  |               |                |
  |<-PeerJoined---|                |
  |               |                |
  |-----Offer---->|                |
  |               |-----Offer----->|
  |               |                |
  |               |<----Answer-----|
  |<----Answer----|                |
  |               |                |
  |--Candidate--->|                |
  |               |---Candidate--->|
  |               |                |
  |               |<---Candidate---|
  |<--Candidate---|                |
```

### Peer Leaving

```
Client A        Server          Client B
  |               |                |
  |               |<---Leave-------|
  |               |                |
  |<--PeerLeft----|                |
  |               |                |
  |  (close connection)            |
```

## Error Handling

### Connection Errors

- **WebSocket Closed**: Client should attempt reconnection
- **Invalid Message**: Server logs error, may send error message
- **Session Not Found**: Server sends error message

### Best Practices

1. **Reconnection**: Implement exponential backoff for reconnection
2. **Timeouts**: Set reasonable timeouts for message responses
3. **Validation**: Validate all incoming messages
4. **Error Messages**: Always check for error type messages

## Implementation Notes

### Server

- Maintains map of sessions to connected clients
- Broadcasts messages to all peers in session except sender
- Automatically removes disconnected clients
- Deletes empty sessions

### Client

- Event-based message handlers
- Non-blocking message sending
- Graceful disconnection on application close
- Sends leave message before disconnecting

## Security Considerations

### Current Implementation

- No authentication
- No message validation beyond JSON parsing
- No rate limiting

### Recommended Enhancements

1. **Authentication**: Verify user identity before joining
2. **Session Passwords**: Optional password protection
3. **Message Validation**: Validate message structure and content
4. **Rate Limiting**: Prevent message spam
5. **Encryption**: Use WSS (WebSocket Secure)

## Testing

### Manual Testing

Use `websocat` or similar tool:

```bash
websocat ws://localhost:8080/ws
```

Send join message:
```json
{"type":"join","session_id":"test-session","peer_id":"test-peer","username":"TestUser"}
```

### Automated Testing

- Unit tests for message serialization/deserialization
- Integration tests for server message routing
- Load tests for concurrent connections
