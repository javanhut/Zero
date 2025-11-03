package signaling

import (
	"encoding/json"

	"github.com/pion/webrtc/v4"
)

type MessageType string

const (
	MessageTypeJoin       MessageType = "join"
	MessageTypeLeave      MessageType = "leave"
	MessageTypeOffer      MessageType = "offer"
	MessageTypeAnswer     MessageType = "answer"
	MessageTypeCandidate  MessageType = "candidate"
	MessageTypePeerJoined MessageType = "peer_joined"
	MessageTypePeerLeft   MessageType = "peer_left"
	MessageTypeError      MessageType = "error"
)

type SignalingMessage struct {
	Type      MessageType     `json:"type"`
	SessionID string          `json:"session_id"`
	PeerID    string          `json:"peer_id"`
	Username  string          `json:"username,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type JoinPayload struct {
	Username string `json:"username"`
}

type PeerJoinedPayload struct {
	PeerID   string `json:"peer_id"`
	Username string `json:"username"`
}

type PeerLeftPayload struct {
	PeerID string `json:"peer_id"`
}

type OfferPayload struct {
	SDP webrtc.SessionDescription `json:"sdp"`
}

type AnswerPayload struct {
	SDP webrtc.SessionDescription `json:"sdp"`
}

type CandidatePayload struct {
	Candidate webrtc.ICECandidateInit `json:"candidate"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

func NewJoinMessage(sessionID, peerID, username string) (*SignalingMessage, error) {
	payload, err := json.Marshal(JoinPayload{Username: username})
	if err != nil {
		return nil, err
	}
	return &SignalingMessage{
		Type:      MessageTypeJoin,
		SessionID: sessionID,
		PeerID:    peerID,
		Username:  username,
		Payload:   payload,
	}, nil
}

func NewLeaveMessage(sessionID, peerID string) *SignalingMessage {
	return &SignalingMessage{
		Type:      MessageTypeLeave,
		SessionID: sessionID,
		PeerID:    peerID,
	}
}

func NewOfferMessage(sessionID, peerID string, sdp webrtc.SessionDescription) (*SignalingMessage, error) {
	payload, err := json.Marshal(OfferPayload{SDP: sdp})
	if err != nil {
		return nil, err
	}
	return &SignalingMessage{
		Type:      MessageTypeOffer,
		SessionID: sessionID,
		PeerID:    peerID,
		Payload:   payload,
	}, nil
}

func NewAnswerMessage(sessionID, peerID string, sdp webrtc.SessionDescription) (*SignalingMessage, error) {
	payload, err := json.Marshal(AnswerPayload{SDP: sdp})
	if err != nil {
		return nil, err
	}
	return &SignalingMessage{
		Type:      MessageTypeAnswer,
		SessionID: sessionID,
		PeerID:    peerID,
		Payload:   payload,
	}, nil
}

func NewCandidateMessage(sessionID, peerID string, candidate webrtc.ICECandidateInit) (*SignalingMessage, error) {
	payload, err := json.Marshal(CandidatePayload{Candidate: candidate})
	if err != nil {
		return nil, err
	}
	return &SignalingMessage{
		Type:      MessageTypeCandidate,
		SessionID: sessionID,
		PeerID:    peerID,
		Payload:   payload,
	}, nil
}

func NewErrorMessage(sessionID, peerID, message string) (*SignalingMessage, error) {
	payload, err := json.Marshal(ErrorPayload{Message: message})
	if err != nil {
		return nil, err
	}
	return &SignalingMessage{
		Type:      MessageTypeError,
		SessionID: sessionID,
		PeerID:    peerID,
		Payload:   payload,
	}, nil
}
