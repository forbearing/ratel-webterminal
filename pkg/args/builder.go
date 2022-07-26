package args

import (
	"net"
	"sync"
)

var builder = &holderBuilder{holder: ratelHolder}

// Used to build argument holder structure. It is private to make sure that
// only 1 instance can be created that modifies singleton instance of argument holder.
type holderBuilder struct {
	holder *holder
	l      sync.Mutex
}

// SetPort sets '--port' argument of ratel-webterminal binary.
func (h *holderBuilder) SetPort(port int) *holderBuilder {
	h.l.Lock()
	defer h.l.Unlock()
	h.holder.port = port
	return h
}

// SetBindAddress sets '--bind-address' argument of ratel-webterminal binary.
func (h *holderBuilder) SetBindAddress(bindAddress net.IP) *holderBuilder {
	h.l.Lock()
	defer h.l.Unlock()
	h.holder.bindAddress = bindAddress
	return h
}

// SetKubeConfigFile sets '--kubeconfig' argument of ratel-webterminal binary.
func (h *holderBuilder) SetKubeConfigFile(kubeConfigFile string) *holderBuilder {
	h.l.Lock()
	defer h.l.Unlock()
	h.holder.kubeConfigFile = kubeConfigFile
	return h
}

// SetLogLevel sets '--log-level' argument of ratel-webterminal binary.
func (h *holderBuilder) SetLogLevel(logLevel string) *holderBuilder {
	h.l.Lock()
	defer h.l.Unlock()
	h.holder.logLevel = logLevel
	return h
}

// SetLogFormat sets '--log-format' argument of ratel-webterminal binary.
func (h *holderBuilder) SetLogFormat(logFormat string) *holderBuilder {
	h.l.Lock()
	defer h.l.Unlock()
	h.holder.logFormat = logFormat
	return h
}

// SetLogLevel sets '--log-format' argument of ratel-webterminal binary.
func (h *holderBuilder) SetLogFile(logFile string) *holderBuilder {
	h.l.Lock()
	defer h.l.Unlock()
	h.holder.logFile = logFile
	return h
}

// NewBuilder returns singleton instance of holder builder.
func NewBuilder() *holderBuilder {
	return builder
}
