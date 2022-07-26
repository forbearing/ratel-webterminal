package session

import (
	"encoding/json"
	"fmt"

	"k8s.io/client-go/tools/remotecommand"
)

// TerminalSize handles pty->process resize events
// Called in a loop from remotecommand as long as the process is running.
func (t TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.SizeCh:
		return &size
	case <-t.doneCh:
		return nil
	}
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remotecommand as long as process is running
func (t TerminalSession) Read(p []byte) (int, error) {
	m, err := t.SockJSSession.Recv()
	if err != nil {
		// Send terminalted signal to process to avoid resource leak.
		return copy(p, END_OF_TRANSMISSION), err
	}

	var msg TerminalMessage
	if err := json.Unmarshal([]byte(m), &msg); err != nil {
		return copy(p, END_OF_TRANSMISSION), err
	}

	switch msg.Op {
	case "stdin":
		return copy(p, msg.Data), nil
	case "resize":
		t.SizeCh <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (t TerminalSession) Write(p []byte) (int, error) {
	msg, err := json.Marshal(TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}

	if err = t.SockJSSession.Send(string(msg)); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t TerminalSession) Toast(p string) error {
	msg, err := json.Marshal(TerminalMessage{
		Op:   "toast",
		Data: p,
	})
	if err != nil {
		return err
	}
	if err = t.SockJSSession.Send(string(msg)); err != nil {
		return err
	}
	return nil
}
