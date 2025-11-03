package signaling

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type MessageHandler func(*SignalingMessage)

type Client struct {
	conn              *websocket.Conn
	serverURL         string
	sessionID         string
	peerID            string
	username          string
	mu                sync.RWMutex
	messageHandlers   map[MessageType][]MessageHandler
	reconnectInterval time.Duration
	done              chan struct{}
	connected         bool
}

func NewClient(serverURL, sessionID, peerID, username string) *Client {
	return &Client{
		serverURL:         serverURL,
		sessionID:         sessionID,
		peerID:            peerID,
		username:          username,
		messageHandlers:   make(map[MessageType][]MessageHandler),
		reconnectInterval: 5 * time.Second,
		done:              make(chan struct{}),
		connected:         false,
	}
}

func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to signaling server: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.mu.Unlock()

	log.Printf("Connected to signaling server: %s", c.serverURL)

	go c.readMessages()

	if err := c.sendJoin(); err != nil {
		return fmt.Errorf("failed to send join message: %w", err)
	}

	return nil
}

func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return
	}

	close(c.done)
	c.sendLeave()

	if c.conn != nil {
		c.conn.Close()
	}

	c.connected = false
	log.Println("Disconnected from signaling server")
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *Client) On(msgType MessageType, handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messageHandlers[msgType] = append(c.messageHandlers[msgType], handler)
}

func (c *Client) sendJoin() error {
	msg, err := NewJoinMessage(c.sessionID, c.peerID, c.username)
	if err != nil {
		return err
	}
	return c.SendMessage(msg)
}

func (c *Client) sendLeave() error {
	msg := NewLeaveMessage(c.sessionID, c.peerID)
	return c.SendMessage(msg)
}

func (c *Client) SendOffer(sdp webrtc.SessionDescription) error {
	msg, err := NewOfferMessage(c.sessionID, c.peerID, sdp)
	if err != nil {
		return err
	}
	return c.SendMessage(msg)
}

func (c *Client) SendAnswer(sdp webrtc.SessionDescription) error {
	msg, err := NewAnswerMessage(c.sessionID, c.peerID, sdp)
	if err != nil {
		return err
	}
	return c.SendMessage(msg)
}

func (c *Client) SendCandidate(candidate webrtc.ICECandidateInit) error {
	msg, err := NewCandidateMessage(c.sessionID, c.peerID, candidate)
	if err != nil {
		return err
	}
	return c.SendMessage(msg)
}

func (c *Client) SendMessage(msg *SignalingMessage) error {
	c.mu.RLock()
	conn := c.conn
	connected := c.connected
	c.mu.RUnlock()

	if !connected || conn == nil {
		return fmt.Errorf("not connected to signaling server")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (c *Client) readMessages() {
	defer func() {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.done:
			return
		default:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			var msg SignalingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			c.handleMessage(&msg)
		}
	}
}

func (c *Client) handleMessage(msg *SignalingMessage) {
	c.mu.RLock()
	handlers, exists := c.messageHandlers[msg.Type]
	c.mu.RUnlock()

	if !exists {
		return
	}

	for _, handler := range handlers {
		go handler(msg)
	}
}

func (c *Client) GetSessionID() string {
	return c.sessionID
}

func (c *Client) GetPeerID() string {
	return c.peerID
}

func (c *Client) GetUsername() string {
	return c.username
}
