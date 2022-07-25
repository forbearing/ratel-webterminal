package websocket

import (
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
)

const END_OF_TRANSMISSION = "\u0004"

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	remotecommand.TerminalSizeQueue
	io.Reader
	io.Writer
}

// TerminalSession implements PtyHandler
type TerminalSession struct {
	conn   *websocket.Conn
	sizeCh chan remotecommand.TerminalSize
	doneCh chan struct{}
}

// TerminalMessage is the messaging protocol between ShellController and TerminalSession.
type TerminalMessage struct {
	Op   string `json:"op"`
	Data string `json:"data"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

var upgrader = func() websocket.Upgrader {
	upgrader := websocket.Upgrader{}
	upgrader.HandshakeTimeout = time.Second * 2
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	return upgrader
}()

type Logger struct {
	conn *websocket.Conn
}
