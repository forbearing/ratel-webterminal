package session

import (
	"sync"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"k8s.io/client-go/tools/remotecommand"
)

const END_OF_TRANSMISSION = "\u0004"

// TerminalMessage is the messaging protocol between ShellController and TerminalSession
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// bind    fe->be     SessionID      Id sent back from TerminalResponse
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type TerminalMessage struct {
	Op, Data, SessionID string
	Rows, Cols          uint16
}

// TerminalSession implements PytHandler (using a SockJS connection)
// The method Next() is used to get the next terminal events.
// The method Read() is used to read message from terminal keystrokes/past buffer.
// The method Wire() is used to
type TerminalSession struct {
	ID            string
	Bound         chan error
	SockJSSession sockjs.Session
	SizeCh        chan remotecommand.TerminalSize
	doneCh        chan struct{}
}

// SessionMap stores a map of all TerminalSession objects and a lock to avoid
// concurrent conflict
type SessionMap struct {
	Sessions map[string]TerminalSession
	l        sync.RWMutex
}
