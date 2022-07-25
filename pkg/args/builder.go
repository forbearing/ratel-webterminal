package args

import "net"

var builder = &holderBuilder{holder: Holder}

// Used to build argument holder structure. It is private to make sure that
// only 1 instance can be created that modifies singleton instance of argument holder.
type holderBuilder struct {
	holder *holder
}

// SetPort sets '--port' argument of ratel-webterminal binary.
func (h *holderBuilder) SetPort(port int) *holderBuilder {
	h.holder.port = port
	return h
}

// SetBindAddress sets '--bind-address' argument of ratel-webterminal binary.
func (h *holderBuilder) SetBindAddress(bindAddress net.IP) *holderBuilder {
	h.holder.bindAddress = bindAddress
	return h
}

// SetKubeConfigFile sets '--kubeconfig' argument of ratel-webterminal binary.
func (h *holderBuilder) SetKubeConfigFile(kubeConfigFile string) *holderBuilder {
	h.holder.kubeConfigFile = kubeConfigFile
	return h
}

// SetLogLevel sets '--log-level' argument of ratel-webterminal binary.
func (h *holderBuilder) SetLogLevel(logLevel string) *holderBuilder {
	h.holder.logLevel = logLevel
	return h
}

// SetLogFormat sets '--log-format' argument of ratel-webterminal binary.
func (h *holderBuilder) SetLogFormat(logFormat string) *holderBuilder {
	h.holder.logFormat = logFormat
	return h
}

// SetLogLevel sets '--log-format' argument of ratel-webterminal binary.
func (h *holderBuilder) SetLogFile(logFile string) *holderBuilder {
	h.holder.logFile = logFile
	return h
}

// NewHolderBuilder returns singleton instance of holder builder.
func NewHolderBuilder() *holderBuilder {
	return builder
}
