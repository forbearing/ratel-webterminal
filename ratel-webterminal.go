package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"

	_ "net/http/pprof"

	"github.com/forbearing/ratel-webterminal/pkg/args"
	"github.com/forbearing/ratel-webterminal/pkg/controller"
	"github.com/forbearing/ratel-webterminal/pkg/logger"
	"github.com/forbearing/ratel-webterminal/pkg/probe"
	"github.com/forbearing/ratel-webterminal/pkg/terminal/websocket"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	argPort           = pflag.Int("port", 8080, "port to listen to for incoming HTTP requests")
	argBindAddress    = pflag.IP("bind-address", net.IPv4(0, 0, 0, 0), "IP address on which to serve the --port, set to 0.0.0.0 for all interfaces by default")
	argKubeConfigFile = pflag.String("kubeconfig", "", "path to kubeconfig file with authorization and master location information")
	argLogLevel       = pflag.String("log-level", "INFO", "level of API request logging, should be one of   'ERROR', 'WARNING|WARN', 'INFO', 'DEBUG' or 'TRACE'")
	argLogFormat      = pflag.String("log-format", "TEXT", "specify log format, should be on of 'TEXT' or 'JSON'")
	argLogFile        = pflag.String("log-output", "/dev/stdout", "specify log file, default output log to /dev/stdout")

	// The flag "--conf" is used to specify a file path, which contains the
	// configuration about how ratel-webterminal to start/bootstrap, such as
	// listen port, bind address, log level, etc.

	// The priority of configuration read from config file takes precedence over
	// the the flags passed from stdin.
	//argConfFile = pflag.String("conf", "", "path to configuration file which the ratel-webterminal will load, config file currently only support yaml format")

)

func init() {
	// flag.CommandLine is the default set of command-line flags, parsed from os.Args.
	// pflag.CommandLine.AddGoFlagSet will add the given *flag.FlagSet to the pflag.FlagSet
	// If you don't know the underlying code does, just ignore it.
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	builder := args.NewBuilder()
	builder.SetPort(*argPort)
	builder.SetBindAddress(*argBindAddress)
	builder.SetKubeConfigFile(*argKubeConfigFile)
	builder.SetLogLevel(*argLogLevel)
	builder.SetLogFormat(*argLogFormat)
	builder.SetLogFile(*argLogFile)
}

func main() {
	logger.Init()
	controller.Init()

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/"))))
	router.HandleFunc("/terminal", websocket.HandleTerminal)
	router.HandleFunc("/logs", websocket.HandleLogs)
	router.HandleFunc("/ws/{namespace}/{pod}/{container}/shell", websocket.HandleWsTerminal)
	router.HandleFunc("/ws/{namespace}/{pod}/{container}/logs", websocket.HandleWsLogs)
	router.HandleFunc("/-/healthy", probe.HandleHealthyProbe)
	router.HandleFunc("/-/ready", probe.HandleReadyProbe)
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	log.Info("Start ratel-webterminal...")
	addr := fmt.Sprintf("%s:%d", args.GetBindAddress(), args.GetPort())
	log.Infof("Listen on %v:%d", args.GetBindAddress(), args.GetPort())
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
