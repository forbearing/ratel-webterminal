package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Write will writes message by call websocket.WriteMessage.
func (l *Logger) Write(p []byte) (int, error) {
	var err error
	if err = l.conn.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close will close websocket connection.
func (l *Logger) Close() error {
	return l.conn.Close()
}

// NewLogger will creates a websocket logger.
func NewLogger(w http.ResponseWriter, r *http.Request, respHeader http.Header) (*Logger, error) {
	conn, err := upgrader.Upgrade(w, r, respHeader)
	if err != nil {
		return nil, err
	}
	return &Logger{
		conn: conn,
	}, nil
}
