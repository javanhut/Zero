# Video Resolution Guide

This document describes the video resolution options available in Zero and how to use them effectively.

## Available Resolutions

Zero supports four video resolution modes:

### SD (Standard Definition)
- **Resolution**: 640x480 pixels
- **Aspect Ratio**: 4:3
- **Bitrate**: ~500 Kbps recommended
- **Use Cases**: 
  - Slow or unstable network connections
  - Low-end devices or limited CPU
  - Battery conservation on laptops
  - Large group calls (5+ participants)

### HD (High Definition)
- **Resolution**: 1280x720 pixels
- **Aspect Ratio**: 16:9
- **Bitrate**: ~1.5 Mbps recommended
- **Use Cases**:
  - Default setting for most scenarios
  - Good balance of quality and performance
  - Standard video conferencing
  - 2-4 participant calls

### Full HD (Full High Definition)
- **Resolution**: 1920x1080 pixels
- **Aspect Ratio**: 16:9
- **Bitrate**: ~2.5 Mbps recommended
- **Use Cases**:
  - High-quality video calls
  - Screen sharing with detail
  - Professional presentations
  - Good network connections

### QHD (Quad High Definition)
- **Resolution**: 2560x1440 pixels
- **Aspect Ratio**: 16:9
- **Bitrate**: ~4-5 Mbps recommended
- **Use Cases**:
  - Maximum quality requirements
  - High-bandwidth LAN environments
  - Professional broadcasting
  - 1-on-1 calls with excellent network

## How to Change Resolution

### During Session Setup

1. Start a new session or connect to an existing one
2. Once the video window opens, locate the Resolution dropdown
3. Select your desired resolution from: SD, HD, Full HD, or QHD
4. The video stream will automatically restart with the new resolution

### Dynamic Resolution Switching

You can change resolution at any time during an active session:

1. Click the Resolution dropdown
2. Select a new resolution
3. The camera will restart automatically with the new settings
4. Your connection to other peers will be maintained
5. The new resolution will be broadcast to all connected participants

## Performance Considerations

### Network Bandwidth

Recommended minimum upload/download speeds per resolution:

| Resolution | Min Upload | Min Download | Recommended |
|------------|------------|--------------|-------------|
| SD         | 0.5 Mbps   | 0.5 Mbps     | 1 Mbps      |
| HD         | 1.5 Mbps   | 1.5 Mbps     | 3 Mbps      |
| Full HD    | 2.5 Mbps   | 2.5 Mbps     | 5 Mbps      |
| QHD        | 4 Mbps     | 4 Mbps       | 8 Mbps      |

### CPU Usage

Higher resolutions require more processing power:

- **SD**: Low CPU usage (5-10%)
- **HD**: Moderate CPU usage (10-20%)
- **Full HD**: High CPU usage (20-35%)
- **QHD**: Very high CPU usage (35-50%+)

### Multi-Peer Scenarios

When connecting with multiple peers, bandwidth requirements multiply:

- 2 peers (1-on-1): Use recommended settings
- 3-4 peers: Consider using HD or lower
- 5+ peers: Use SD or HD for best performance
- 10+ peers: Use SD to ensure stability

## Troubleshooting

### Video is Choppy or Freezing

**Solutions**:
1. Switch to a lower resolution (HD to SD)
2. Close bandwidth-heavy applications
3. Use wired ethernet instead of WiFi
4. Reduce number of participants

### High CPU Usage

**Solutions**:
1. Switch to SD or HD resolution
2. Close unnecessary applications
3. Ensure proper system cooling
4. Update graphics drivers

### Poor Video Quality Despite High Resolution

**Possible Issues**:
- Network congestion causing packet loss
- Camera hardware limitations
- Insufficient lighting
- Codec compression artifacts

**Solutions**:
1. Test camera in other applications
2. Improve lighting conditions
3. Check network statistics with the Stats button
4. Try a different camera with Select Camera

### Resolution Change Not Working

**Solutions**:
1. Ensure camera is not in use by other applications
2. Check camera permissions
3. Verify camera supports the selected resolution
4. Review logs for error messages

## Camera Compatibility

Not all cameras support all resolutions:

- Most modern webcams support SD and HD
- Higher-end webcams support Full HD
- Professional cameras typically support QHD
- Built-in laptop cameras vary by model

If a resolution is not supported by your camera, Zero will attempt to fall back to a supported resolution automatically.

## Configuration File Settings

The default resolution can be set in `config.yaml`:

```yaml
media:
  video:
    codec: "vp8"
    resolution: "HD"  # Options: SD, HD, FullHD, QHD
    bitrate: 1500000
```

## Best Practices

1. **Start with HD**: Default setting works for most scenarios
2. **Test your network**: Use Stats button to monitor performance
3. **Adjust based on participants**: More peers = lower resolution
4. **Consider your use case**: Presentations need higher quality than casual calls
5. **Monitor CPU**: Keep usage under 50% for smooth performance
6. **WiFi users**: Use SD or HD for more stable connections
7. **LAN/Ethernet users**: Can use Full HD or QHD safely

## Technical Details

### Video Encoding

All resolutions use the VP8 codec for maximum compatibility:
- Keyframe interval: 10 seconds
- Bitrate adaptation: Enabled
- Hardware acceleration: When available

### Resolution Switching Process

When you change resolution:

1. Current video stream is stopped gracefully
2. Camera is released
3. New stream is initialized with selected resolution
4. WebRTC tracks are recreated
5. Peers are updated with new media capabilities
6. Video transmission resumes

The entire process typically takes 1-2 seconds.

## API Integration

For developers integrating resolution changes programmatically:

```go
// Available resolutions
resolutions := []string{"SD", "HD", "FullHD", "QHD"}

// Start stream with specific resolution
stream, err := camera.StartVideoStream("HD", "", updateFunc)

// Get current resolution from stats
stats := videoStream.GetStats()
currentResolution := stats.Resolution
```

## Future Enhancements

Planned improvements for resolution handling:

- Auto-resolution based on network conditions
- Custom resolution support
- Per-peer resolution negotiation
- Bandwidth estimation and adaptation
- Hardware acceleration preferences

For questions or issues, consult the main documentation or open an issue on GitHub.
