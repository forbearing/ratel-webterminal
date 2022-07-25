package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/remotecommand"
)

// NewTerminalSessionWs create TerminalSession
func NewTerminalSessionWs(conn *websocket.Conn) *TerminalSession {
	return &TerminalSession{
		conn:   conn,
		sizeCh: make(chan remotecommand.TerminalSize),
		doneCh: make(chan struct{}),
	}
}

// NewTerminalSession create TerminalSession
func NewTerminalSession(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*TerminalSession, error) {
	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	session := &TerminalSession{
		conn:   conn,
		sizeCh: make(chan remotecommand.TerminalSize),
		doneCh: make(chan struct{}),
	}
	return session, nil
}

// Next called in a loop from remotecommand as long as the process is running
func (t *TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeCh:
		return &size
	case <-t.doneCh:
		return nil
	}
}

// Read called in a loop from remotecommand as long as the process is running
func (t *TerminalSession) Read(p []byte) (int, error) {
	_, message, err := t.conn.ReadMessage()
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			log.Println("closed network connection")
			return copy(p, END_OF_TRANSMISSION), nil
		} else {
			log.Printf("read message err: %v", err)
			return copy(p, END_OF_TRANSMISSION), err
		}
		//return copy(p, EndOfTransmission), err
	}
	var msg TerminalMessage
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		log.Printf("read parse message err: %v", err)
		// return 0, nil
		return copy(p, END_OF_TRANSMISSION), err
	}
	switch msg.Op {
	case "stdin":
		return copy(p, msg.Data), nil
	case "resize":
		t.sizeCh <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		log.Printf("unknown message type '%s'", msg.Op)
		// return 0, nil
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// Write called from remotecommand whenever there is any output
func (t *TerminalSession) Write(p []byte) (int, error) {
	msg, err := json.Marshal(TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	})
	if err != nil {
		log.Printf("write parse message err: %v", err)
		return 0, err
	}
	if err := t.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Printf("write message err: %v", err)
		return 0, err
	}
	return len(p), nil
}

// Close close session
func (t *TerminalSession) Close() error {
	close(t.doneCh)
	return t.conn.Close()
}
