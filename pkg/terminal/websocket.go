package terminal

import (
	"context"
	"net/http"

	"github.com/forbearing/k8s/pod"
	"github.com/forbearing/ratel-webterminal/pkg/args"
	"github.com/forbearing/ratel-webterminal/pkg/terminal/websocket"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// HandleTerminal handle "/terminal" connections.
func HandleTerminal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Error("HandleTerminal error: ", http.StatusText(http.StatusMethodNotAllowed))
		http.Error(w, "HandleTerminal: "+http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
	http.ServeFile(w, r, "./frontend/terminal.html")
}

// HandleLogs handle "/logs" connections.
func HandleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Error("HandleLogs error: ", http.StatusText(http.StatusMethodNotAllowed))
		http.Error(w, "HandleLogs: "+http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
	http.ServeFile(w, r, "./frontend/logs.html")
}

// HandleWebsocketTerminal
func HandleWsTerminal(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container"]
	log.Infof("exec pod: %s/%s, container: %s", namespace, podName, containerName)
	log.Info(r.URL)

	pty, err := websocket.NewTerminalSession(w, r, nil)
	if err != nil {
		log.Error("create terminal session error: ", err)
		return
	}
	defer func() {
		log.Info("close terminal session")
		pty.Close()
	}()

	podHandler, err := pod.New(context.TODO(), args.Holder.GetKubeConfigFile(), "")
	err = podHandler.Execute(podName, containerName, []string{"bash"}, pty)
	if err != nil {
		log.Error("create pod shell error: ", err)
		if err = podHandler.Execute(podName, containerName, []string{"sh"}, pty); err != nil {
			log.Error("create pod shell error: ", err)
		}
	}
}

// HandleWebsocketLogs
func HandleWsLogs(w http.ResponseWriter, r *http.Request) {}
