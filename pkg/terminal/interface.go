package terminal

import (
	"io"

	"k8s.io/client-go/tools/remotecommand"
)

//type PtyHandler interface {
//    io.Reader
//    io.Writer
//    remotecommand.TerminalSizeQueue
//}

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	remotecommand.TerminalSizeQueue
	Done()
	Tty() bool
	Stdin() io.Reader
	Stdout() io.Writer
	Stderr() io.Writer
}
