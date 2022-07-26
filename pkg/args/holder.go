package args

import (
	"net"
)

var ratelHolder = &holder{}

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
func GetPort() int {
	return ratelHolder.port
}

// GetBindAddress returns "--bind-address" argument of ratel-webterminal binary.
func GetBindAddress() net.IP {
	return ratelHolder.bindAddress
}

// GetKubeConfigFile returns "--kubeconfig" argument of ratel-webterminal binary.
func GetKubeConfigFile() string {
	return ratelHolder.kubeConfigFile
}

// GetLogLevel returns "--log-level" argument of ratel-webterminal binary.
func GetLogLevel() string {
	return ratelHolder.logLevel
}

// GetLogFormat returns "--log-format" argument of ratel-webterminal binary.
func GetLogFormat() string {
	return ratelHolder.logFormat
}

// GetLogFile returns "--log-file" argument of ratel-webterminal binary.
func GetLogFile() string {
	return ratelHolder.logFile
}
