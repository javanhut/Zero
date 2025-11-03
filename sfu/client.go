package sfu

import (
	"fmt"
	"log"

	"github.com/pion/webrtc/v4"
)

type Client struct {
	sfuURL       string
	sessionID    string
	peerID       string
	pc           *webrtc.PeerConnection
	localTracks  []*webrtc.TrackLocalStaticSample
	remoteTracks []*webrtc.TrackRemote
	onTrack      func(*webrtc.TrackRemote)
}

type ClientConfig struct {
	SFUURL    string
	SessionID string
	PeerID    string
	OnTrack   func(*webrtc.TrackRemote)
}

func NewClient(config ClientConfig) (*Client, error) {
	log.Printf("Creating SFU client for session %s", config.SessionID)

	return &Client{
		sfuURL:       config.SFUURL,
		sessionID:    config.SessionID,
		peerID:       config.PeerID,
		localTracks:  make([]*webrtc.TrackLocalStaticSample, 0),
		remoteTracks: make([]*webrtc.TrackRemote, 0),
		onTrack:      config.OnTrack,
	}, nil
}

func (c *Client) Connect() error {
	log.Printf("Connecting to SFU server at %s", c.sfuURL)
	return fmt.Errorf("ION SFU integration not yet implemented - use peer-to-peer mode")
}

func (c *Client) AddTrack(track *webrtc.TrackLocalStaticSample) error {
	c.localTracks = append(c.localTracks, track)
	log.Printf("Added track to SFU client: %s", track.ID())
	return nil
}

func (c *Client) Close() error {
	if c.pc != nil {
		return c.pc.Close()
	}
	return nil
}
