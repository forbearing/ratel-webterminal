package session

import (
	"crypto/rand"
	"encoding/hex"

	log "github.com/sirupsen/logrus"
)

// Get return a given TerminalSession by sessionID.
func (s *SessionMap) Get(sessionID string) TerminalSession {
	s.l.RLock()
	defer s.l.RUnlock()
	return s.Sessions[sessionID]
}

// Set store a TerminalSession to SessionMap.
func (s *SessionMap) Set(sessionID string, session TerminalSession) {
	s.l.Lock()
	defer s.l.Unlock()
	s.Sessions[sessionID] = session
}

// Close shuts down the SockJs connection and sends the status code and reason
// to the client.
// Can happen if the process exists or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (s *SessionMap) Close(sessionID string, status uint32, reason string) {
	s.l.Lock()
	defer s.l.Unlock()
	err := s.Sessions[sessionID].SockJSSession.Close(status, reason)
	if err != nil {
		log.Error("close sockJS session error: ", err)
	}
	delete(s.Sessions, sessionID)
}

// GenTerminalSessionID generates a random session ID string. the format is not
// ready interesting.
// This ID is used to identify the session when the client opens the SockJs connection.
// Not the same as the SockJs session ID! We cann't use that as that is generated
// on the client side and we don't have it yet at this point.
func GenTerminalSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}
