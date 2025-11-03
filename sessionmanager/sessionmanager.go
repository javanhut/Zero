package sessionmanager

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Peer struct {
	PeerID    string
	Username  string
	Connected bool
	JoinedAt  time.Time
}

type SessionInfo struct {
	SessionID string
	Peers     map[string]*Peer
	CreatedAt time.Time
	Active    int
}

type SessionManager struct {
	sessions map[string]*SessionInfo
	mu       sync.RWMutex
}

func New() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SessionInfo),
	}
}

func (sm *SessionManager) CheckForSession(sessionID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	_, ok := sm.sessions[sessionID]
	if !ok {
		log.Println("Session doesn't exist in table")
		return false
	}
	sessionExistStr := fmt.Sprintf("Session %s exists in Session table", sessionID)
	log.Println(sessionExistStr)
	return true
}

func (sm *SessionManager) GetUsername(sessionID string) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return "Unknown User"
	}

	for _, peer := range session.Peers {
		return peer.Username
	}
	return "Unknown User"
}

func (sm *SessionManager) CreateNewSession() (string, string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID := uuid.New().String()
	peerID := uuid.New().String()
	username := fmt.Sprintf("User_%s", peerID[:8])

	session := &SessionInfo{
		SessionID: sessionID,
		Peers:     make(map[string]*Peer),
		CreatedAt: time.Now(),
		Active:    1,
	}

	session.Peers[peerID] = &Peer{
		PeerID:    peerID,
		Username:  username,
		Connected: true,
		JoinedAt:  time.Now(),
	}

	sm.sessions[sessionID] = session
	sessionString := fmt.Sprintf("Created new session with id: %s for user: %s (peer: %s)", sessionID, username, peerID)
	log.Println(sessionString)

	return sessionID, username
}

func (sm *SessionManager) JoinSession(sessionID string) (string, string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return "", "", fmt.Errorf("session %s not found", sessionID)
	}

	peerID := uuid.New().String()
	username := fmt.Sprintf("User_%s", peerID[:8])

	session.Peers[peerID] = &Peer{
		PeerID:    peerID,
		Username:  username,
		Connected: true,
		JoinedAt:  time.Now(),
	}

	session.Active = len(session.Peers)
	log.Printf("User %s (peer: %s) joined session %s", username, peerID, sessionID)

	return peerID, username, nil
}

func (sm *SessionManager) AddPeerToSession(sessionID, peerID, username string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if _, exists := session.Peers[peerID]; exists {
		return fmt.Errorf("peer %s already in session", peerID)
	}

	session.Peers[peerID] = &Peer{
		PeerID:    peerID,
		Username:  username,
		Connected: true,
		JoinedAt:  time.Now(),
	}

	session.Active = len(session.Peers)
	log.Printf("Added peer %s (%s) to session %s", peerID, username, sessionID)
	return nil
}

func (sm *SessionManager) RemovePeerFromSession(sessionID, peerID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}

	delete(session.Peers, peerID)
	session.Active = len(session.Peers)

	log.Printf("Removed peer %s from session %s", peerID, sessionID)

	if len(session.Peers) == 0 {
		delete(sm.sessions, sessionID)
		log.Printf("Deleted empty session: %s", sessionID)
	}

	return nil
}

func (sm *SessionManager) GetPeersInSession(sessionID string) ([]*Peer, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	peers := make([]*Peer, 0, len(session.Peers))
	for _, peer := range session.Peers {
		peers = append(peers, peer)
	}

	return peers, nil
}

func (sm *SessionManager) GetSession(sessionID string) (*SessionInfo, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, exists := sm.sessions[sessionID]
	if !exists {
		log.Println("Nothing to delete as session is not found")
		return
	}

	delete(sm.sessions, sessionID)
	deletedStr := fmt.Sprintf("Deleting Session [%s] from session table", sessionID)
	log.Println(deletedStr)
}

func (sm *SessionManager) GetAllSessions() []*SessionInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*SessionInfo, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

func (sm *SessionManager) UpdatePeerConnection(sessionID, peerID string, connected bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}

	peer, ok := session.Peers[peerID]
	if !ok {
		return fmt.Errorf("peer %s not found in session", peerID)
	}

	peer.Connected = connected
	return nil
}
