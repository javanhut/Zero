package webrtc

import (
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v4"
)

type PeerConnection struct {
	pc           *webrtc.PeerConnection
	peerID       string
	sessionID    string
	localTracks  []*webrtc.TrackLocalStaticSample
	remoteTracks []*webrtc.TrackRemote
	onTrack      func(*webrtc.TrackRemote, *webrtc.RTPReceiver)
	onDisconnect func(string)
	onICE        func(*webrtc.ICECandidate)
	mu           sync.RWMutex
	connected    bool
}

type PeerConnectionConfig struct {
	PeerID       string
	SessionID    string
	Config       webrtc.Configuration
	OnTrack      func(*webrtc.TrackRemote, *webrtc.RTPReceiver)
	OnDisconnect func(string)
	OnICE        func(*webrtc.ICECandidate)
}

func NewPeerConnection(config PeerConnectionConfig) (*PeerConnection, error) {
	pc, err := webrtc.NewPeerConnection(config.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	peer := &PeerConnection{
		pc:           pc,
		peerID:       config.PeerID,
		sessionID:    config.SessionID,
		localTracks:  make([]*webrtc.TrackLocalStaticSample, 0),
		remoteTracks: make([]*webrtc.TrackRemote, 0),
		onTrack:      config.OnTrack,
		onDisconnect: config.OnDisconnect,
		onICE:        config.OnICE,
		connected:    false,
	}

	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil && peer.onICE != nil {
			peer.onICE(candidate)
		}
	})

	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("Received track from peer %s: %s", peer.peerID, track.ID())
		peer.mu.Lock()
		peer.remoteTracks = append(peer.remoteTracks, track)
		peer.mu.Unlock()

		if peer.onTrack != nil {
			peer.onTrack(track, receiver)
		}
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("Peer %s connection state: %s", peer.peerID, state.String())

		peer.mu.Lock()
		wasConnected := peer.connected
		peer.mu.Unlock()

		switch state {
		case webrtc.PeerConnectionStateConnected:
			peer.mu.Lock()
			peer.connected = true
			peer.mu.Unlock()
			log.Printf("Peer %s connected", peer.peerID)

		case webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateDisconnected, webrtc.PeerConnectionStateClosed:
			peer.mu.Lock()
			peer.connected = false
			peer.mu.Unlock()

			if wasConnected && peer.onDisconnect != nil {
				peer.onDisconnect(peer.peerID)
			}
		}
	})

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("Peer %s ICE connection state: %s", peer.peerID, state.String())
	})

	return peer, nil
}

func (p *PeerConnection) AddTrack(track *webrtc.TrackLocalStaticSample) error {
	sender, err := p.pc.AddTrack(track)
	if err != nil {
		return fmt.Errorf("failed to add track: %w", err)
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := sender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	p.mu.Lock()
	p.localTracks = append(p.localTracks, track)
	p.mu.Unlock()

	log.Printf("Added track to peer %s: %s", p.peerID, track.ID())
	return nil
}

func (p *PeerConnection) CreateOffer() (webrtc.SessionDescription, error) {
	offer, err := p.pc.CreateOffer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("failed to create offer: %w", err)
	}

	if err := p.pc.SetLocalDescription(offer); err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("failed to set local description: %w", err)
	}

	return offer, nil
}

func (p *PeerConnection) CreateAnswer() (webrtc.SessionDescription, error) {
	answer, err := p.pc.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("failed to create answer: %w", err)
	}

	if err := p.pc.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("failed to set local description: %w", err)
	}

	return answer, nil
}

func (p *PeerConnection) SetRemoteDescription(sdp webrtc.SessionDescription) error {
	if err := p.pc.SetRemoteDescription(sdp); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}
	return nil
}

func (p *PeerConnection) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	if err := p.pc.AddICECandidate(candidate); err != nil {
		return fmt.Errorf("failed to add ICE candidate: %w", err)
	}
	return nil
}

func (p *PeerConnection) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connected = false

	if p.pc != nil {
		return p.pc.Close()
	}
	return nil
}

func (p *PeerConnection) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

func (p *PeerConnection) GetPeerID() string {
	return p.peerID
}

func (p *PeerConnection) GetSessionID() string {
	return p.sessionID
}

func (p *PeerConnection) GetRemoteTracks() []*webrtc.TrackRemote {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tracks := make([]*webrtc.TrackRemote, len(p.remoteTracks))
	copy(tracks, p.remoteTracks)
	return tracks
}
