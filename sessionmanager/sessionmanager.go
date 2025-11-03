package sessionmanager

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

type SessionInfo struct {
	Active   int
	Username string
}

type SessionManager struct {
	SessionTable map[string]SessionInfo
}

func (sm *SessionManager) New() {
	var sessionTable = make(map[string]SessionInfo)
	sm.SessionTable = sessionTable

}

func (sm *SessionManager) CheckForSession(session_id string) bool {
	_, ok := sm.SessionTable[session_id]
	if !ok {
		log.Println("Session doesn't exist in table")
		return false
	} else {
		sessionExistStr := fmt.Sprintf("Session %s exists in Session table", session_id)
		log.Println(sessionExistStr)
		return true
	}
}

func (sm *SessionManager) GetUsername(session_id string) string {
	info, ok := sm.SessionTable[session_id]
	if !ok {
		return "Unknown User"
	}
	return info.Username
}

func (sm *SessionManager) CreateNewSession() (string, string) {
	id := uuid.New()
	username := fmt.Sprintf("User_%s", id.String()[:8])
	sm.SessionTable[id.String()] = SessionInfo{Active: 1, Username: username}
	session_string := fmt.Sprintf("Created new session with id: %s for user: %s", id.String(), username)
	log.Println(session_string)
	return id.String(), username
}

func (sm *SessionManager) DeleteSession(session_id string) {
	sessionExists := sm.CheckForSession(session_id)
	if !sessionExists {
		log.Println("Nothing to delete as session is not found")
	} else {
		delete(sm.SessionTable, session_id)
		deletedStr := fmt.Sprintf("Deleting Session [%s] from session table", session_id)
		log.Println(deletedStr)
	}

}
