package args

import (
	"net"
)

var Holder = &holder{}

// holder is a structure contains all arguments values passed to ratel-terminal.
type holder struct {
	port           int
	bindAddress    net.IP
	kubeConfigFile string
	logLevel       string
	logFormat      string
	logFile        string
}

// GetPort returns "--port" argument of ratel-webterminal binary.
func (h *holder) GetPort() int {
	return h.port
}

// GetBindAddress returns "--bind-address" argument of ratel-webterminal binary.
func (h *holder) GetBindAddress() net.IP {
	return h.bindAddress
}

// GetKubeConfigFile returns "--kubeconfig" argument of ratel-webterminal binary.
func (h *holder) GetKubeConfigFile() string {
	return h.kubeConfigFile
}

// GetLogLevel returns "--log-level" argument of ratel-webterminal binary.
func (h *holder) GetLogLevel() string {
	return h.logLevel
}

// GetLogFormat returns "--log-format" argument of ratel-webterminal binary.
func (h *holder) GetLogFormat() string {
	return h.logFormat
}

// GetLogFile returns "--log-file" argument of ratel-webterminal binary.
func (h *holder) GetLogFile() string {
	return h.logFile
}
