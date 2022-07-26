package websocket

import (
	"context"
	"net/http"
	"strconv"

	"github.com/forbearing/k8s/pod"
	"github.com/forbearing/ratel-webterminal/pkg/args"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
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

	pty, err := NewTerminalSession(w, r, nil)
	if err != nil {
		log.Error("create terminal session error: ", err)
		return
	}
	defer func() {
		log.Info("close terminal session")
		pty.Close()
	}()

	podHandler, err := pod.New(context.TODO(), args.Holder.GetKubeConfigFile(), namespace)
	err = podHandler.Execute(podName, containerName, []string{"bash"}, pty)
	if err != nil {
		if err = podHandler.Execute(podName, containerName, []string{"sh"}, pty); err != nil {
			log.Error("create pod shell error: ", err)
		}
	}
}

// HandleWebsocketLogs
func HandleWsLogs(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container"]
	tailLines, _ := strconv.ParseInt(r.URL.Query().Get("tail"), 10, 64)
	log.Infof("get pod logs: %s/%s, container: %s, tailLines: %d\n", namespace, podName, containerName, tailLines)
	log.Info(r.URL)

	writer, err := NewLogger(w, r, nil)
	if err != nil {
		log.Error("websocket.NewLogger error: ", err)
		return
	}
	defer func() {
		log.Println("close logger session.")
		writer.Close()
	}()

	if tailLines == 0 {
		tailLines = 50
	}
	logOptions := pod.LogOptions{
		PodLogOptions: corev1.PodLogOptions{
			Follow:    true,
			TailLines: &tailLines,
		},
		Writer: writer,
	}
	podHandler, err := pod.New(context.TODO(), args.Holder.GetKubeConfigFile(), namespace)
	if err != nil {
		log.Error("get pod handler error")
		return
	}
	if err = podHandler.Log(podName, &logOptions); err != nil {
		log.Error("get pod log error")
	}
}
