package signaling

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type ServerClient struct {
	conn      *websocket.Conn
	sessionID string
	peerID    string
	username  string
	send      chan []byte
}

type Session struct {
	clients map[string]*ServerClient
	mu      sync.RWMutex
}

type Server struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		sessions: make(map[string]*Session),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &ServerClient{
		conn: conn,
		send: make(chan []byte, 256),
	}

	go client.writePump()
	go s.readPump(client)
}

func (s *Server) getOrCreateSession(sessionID string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		session = &Session{
			clients: make(map[string]*ServerClient),
		}
		s.sessions[sessionID] = session
		log.Printf("Created new session: %s", sessionID)
	}
	return session
}

func (s *Server) addClientToSession(sessionID string, client *ServerClient) {
	session := s.getOrCreateSession(sessionID)
	session.mu.Lock()
	defer session.mu.Unlock()

	session.clients[client.peerID] = client
	log.Printf("Added client %s to session %s", client.peerID, sessionID)
}

func (s *Server) removeClientFromSession(sessionID, peerID string) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	session.mu.Lock()
	delete(session.clients, peerID)
	clientCount := len(session.clients)
	session.mu.Unlock()

	log.Printf("Removed client %s from session %s", peerID, sessionID)

	if clientCount == 0 {
		s.mu.Lock()
		delete(s.sessions, sessionID)
		s.mu.Unlock()
		log.Printf("Deleted empty session: %s", sessionID)
	}
}

func (s *Server) broadcastToSession(sessionID, senderPeerID string, message []byte) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	for peerID, client := range session.clients {
		if peerID != senderPeerID {
			select {
			case client.send <- message:
			default:
				log.Printf("Failed to send message to client %s", peerID)
			}
		}
	}
}

func (s *Server) notifyPeerJoined(sessionID, newPeerID, username string) {
	payload, _ := json.Marshal(PeerJoinedPayload{
		PeerID:   newPeerID,
		Username: username,
	})

	msg := &SignalingMessage{
		Type:      MessageTypePeerJoined,
		SessionID: sessionID,
		PeerID:    newPeerID,
		Username:  username,
		Payload:   payload,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal peer joined message: %v", err)
		return
	}

	s.broadcastToSession(sessionID, newPeerID, msgBytes)
}

func (s *Server) notifyPeerLeft(sessionID, peerID string) {
	payload, _ := json.Marshal(PeerLeftPayload{
		PeerID: peerID,
	})

	msg := &SignalingMessage{
		Type:      MessageTypePeerLeft,
		SessionID: sessionID,
		PeerID:    peerID,
		Payload:   payload,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal peer left message: %v", err)
		return
	}

	s.broadcastToSession(sessionID, peerID, msgBytes)
}

func (s *Server) readPump(client *ServerClient) {
	defer func() {
		if client.sessionID != "" && client.peerID != "" {
			s.removeClientFromSession(client.sessionID, client.peerID)
			s.notifyPeerLeft(client.sessionID, client.peerID)
		}
		client.conn.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg SignalingMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		s.handleMessage(client, &msg, message)
	}
}

func (s *Server) handleMessage(client *ServerClient, msg *SignalingMessage, rawMsg []byte) {
	switch msg.Type {
	case MessageTypeJoin:
		client.sessionID = msg.SessionID
		client.peerID = msg.PeerID
		client.username = msg.Username
		s.addClientToSession(msg.SessionID, client)
		s.notifyPeerJoined(msg.SessionID, msg.PeerID, msg.Username)
		log.Printf("Client %s joined session %s", msg.PeerID, msg.SessionID)

	case MessageTypeLeave:
		s.removeClientFromSession(msg.SessionID, msg.PeerID)
		s.notifyPeerLeft(msg.SessionID, msg.PeerID)
		log.Printf("Client %s left session %s", msg.PeerID, msg.SessionID)

	case MessageTypeOffer, MessageTypeAnswer, MessageTypeCandidate:
		s.broadcastToSession(msg.SessionID, msg.PeerID, rawMsg)

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (client *ServerClient) writePump() {
	defer func() {
		client.conn.Close()
	}()

	for {
		message, ok := <-client.send
		if !ok {
			client.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to write message: %v", err)
			return
		}
	}
}

func (s *Server) Start(addr string) error {
	http.HandleFunc("/ws", s.HandleWebSocket)
	log.Printf("Signaling server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}
