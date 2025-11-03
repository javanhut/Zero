package webrtc

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/javanhut/zero/signaling"
	"github.com/pion/webrtc/v4"
)

type RemoteTrackHandler func(peerID string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)

type Manager struct {
	peers            map[string]*PeerConnection
	config           *Config
	signaling        *signaling.Client
	localTracks      []*webrtc.TrackLocalStaticSample
	onRemoteTrack    RemoteTrackHandler
	onPeerDisconnect func(peerID string)
	mu               sync.RWMutex
}

type ManagerConfig struct {
	WebRTCConfig     *Config
	SignalingClient  *signaling.Client
	OnRemoteTrack    RemoteTrackHandler
	OnPeerDisconnect func(peerID string)
}

func NewManager(config ManagerConfig) *Manager {
	m := &Manager{
		peers:            make(map[string]*PeerConnection),
		config:           config.WebRTCConfig,
		signaling:        config.SignalingClient,
		localTracks:      make([]*webrtc.TrackLocalStaticSample, 0),
		onRemoteTrack:    config.OnRemoteTrack,
		onPeerDisconnect: config.OnPeerDisconnect,
	}

	m.setupSignalingHandlers()
	return m
}

func (m *Manager) setupSignalingHandlers() {
	m.signaling.On(signaling.MessageTypePeerJoined, m.handlePeerJoined)
	m.signaling.On(signaling.MessageTypePeerLeft, m.handlePeerLeft)
	m.signaling.On(signaling.MessageTypeOffer, m.handleOffer)
	m.signaling.On(signaling.MessageTypeAnswer, m.handleAnswer)
	m.signaling.On(signaling.MessageTypeCandidate, m.handleCandidate)
}

func (m *Manager) handlePeerJoined(msg *signaling.SignalingMessage) {
	var payload signaling.PeerJoinedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal peer joined payload: %v", err)
		return
	}

	if payload.PeerID == m.signaling.GetPeerID() {
		return
	}

	log.Printf("Peer joined: %s (%s)", payload.Username, payload.PeerID)

	if err := m.createPeerConnection(payload.PeerID); err != nil {
		log.Printf("Failed to create peer connection: %v", err)
		return
	}

	if err := m.sendOffer(payload.PeerID); err != nil {
		log.Printf("Failed to send offer: %v", err)
	}
}

func (m *Manager) handlePeerLeft(msg *signaling.SignalingMessage) {
	var payload signaling.PeerLeftPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal peer left payload: %v", err)
		return
	}

	log.Printf("Peer left: %s", payload.PeerID)
	m.removePeer(payload.PeerID)
}

func (m *Manager) handleOffer(msg *signaling.SignalingMessage) {
	var payload signaling.OfferPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal offer payload: %v", err)
		return
	}

	log.Printf("Received offer from peer: %s", msg.PeerID)

	m.mu.RLock()
	peer, exists := m.peers[msg.PeerID]
	m.mu.RUnlock()

	if !exists {
		if err := m.createPeerConnection(msg.PeerID); err != nil {
			log.Printf("Failed to create peer connection: %v", err)
			return
		}
		m.mu.RLock()
		peer = m.peers[msg.PeerID]
		m.mu.RUnlock()
	}

	if err := peer.SetRemoteDescription(payload.SDP); err != nil {
		log.Printf("Failed to set remote description: %v", err)
		return
	}

	if err := m.sendAnswer(msg.PeerID); err != nil {
		log.Printf("Failed to send answer: %v", err)
	}
}

func (m *Manager) handleAnswer(msg *signaling.SignalingMessage) {
	var payload signaling.AnswerPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal answer payload: %v", err)
		return
	}

	log.Printf("Received answer from peer: %s", msg.PeerID)

	m.mu.RLock()
	peer, exists := m.peers[msg.PeerID]
	m.mu.RUnlock()

	if !exists {
		log.Printf("Received answer for unknown peer: %s", msg.PeerID)
		return
	}

	if err := peer.SetRemoteDescription(payload.SDP); err != nil {
		log.Printf("Failed to set remote description: %v", err)
	}
}

func (m *Manager) handleCandidate(msg *signaling.SignalingMessage) {
	var payload signaling.CandidatePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal candidate payload: %v", err)
		return
	}

	m.mu.RLock()
	peer, exists := m.peers[msg.PeerID]
	m.mu.RUnlock()

	if !exists {
		log.Printf("Received candidate for unknown peer: %s", msg.PeerID)
		return
	}

	if err := peer.AddICECandidate(payload.Candidate); err != nil {
		log.Printf("Failed to add ICE candidate: %v", err)
	}
}

func (m *Manager) createPeerConnection(peerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.peers[peerID]; exists {
		return nil
	}

	peer, err := NewPeerConnection(PeerConnectionConfig{
		PeerID:    peerID,
		SessionID: m.signaling.GetSessionID(),
		Config:    m.config.ToWebRTCConfig(),
		OnTrack: func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			if m.onRemoteTrack != nil {
				m.onRemoteTrack(peerID, track, receiver)
			}
		},
		OnDisconnect: func(pid string) {
			m.removePeer(pid)
			if m.onPeerDisconnect != nil {
				m.onPeerDisconnect(pid)
			}
		},
		OnICE: func(candidate *webrtc.ICECandidate) {
			if candidate == nil {
				return
			}
			init := candidate.ToJSON()
			if err := m.signaling.SendCandidate(init); err != nil {
				log.Printf("Failed to send ICE candidate: %v", err)
			}
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create peer connection: %w", err)
	}

	for _, track := range m.localTracks {
		if err := peer.AddTrack(track); err != nil {
			log.Printf("Failed to add local track to peer: %v", err)
		}
	}

	m.peers[peerID] = peer
	log.Printf("Created peer connection for: %s", peerID)
	return nil
}

func (m *Manager) sendOffer(peerID string) error {
	m.mu.RLock()
	peer, exists := m.peers[peerID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	offer, err := peer.CreateOffer()
	if err != nil {
		return err
	}

	return m.signaling.SendOffer(offer)
}

func (m *Manager) sendAnswer(peerID string) error {
	m.mu.RLock()
	peer, exists := m.peers[peerID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	answer, err := peer.CreateAnswer()
	if err != nil {
		return err
	}

	return m.signaling.SendAnswer(answer)
}

func (m *Manager) AddLocalTrack(track *webrtc.TrackLocalStaticSample) error {
	m.mu.Lock()
	m.localTracks = append(m.localTracks, track)
	peers := make([]*PeerConnection, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer)
	}
	m.mu.Unlock()

	for _, peer := range peers {
		if err := peer.AddTrack(track); err != nil {
			log.Printf("Failed to add track to peer %s: %v", peer.GetPeerID(), err)
		}
	}

	log.Printf("Added local track: %s", track.ID())
	return nil
}

func (m *Manager) removePeer(peerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	peer, exists := m.peers[peerID]
	if !exists {
		return
	}

	peer.Close()
	delete(m.peers, peerID)
	log.Printf("Removed peer: %s", peerID)
}

func (m *Manager) GetPeers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peerIDs := make([]string, 0, len(m.peers))
	for peerID := range m.peers {
		peerIDs = append(peerIDs, peerID)
	}
	return peerIDs
}

func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, peer := range m.peers {
		peer.Close()
	}
	m.peers = make(map[string]*PeerConnection)
}
