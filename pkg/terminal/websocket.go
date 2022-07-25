package terminal

import (
	"context"
	"net/http"

	"github.com/forbearing/k8s/pod"
	"github.com/forbearing/ratel-webterminal/pkg/args"
	"github.com/forbearing/ratel-webterminal/pkg/terminal/websocket"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
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

	//err = startProcess(k8s.Clientset(), k8s.RESTConfig(), []string{"bash"}, pty, namespace, podName, containerName)
	//if err != nil {
	//    log.Error("create pod shell error: ", err)
	//}
}

func startProcess(k8sclient kubernetes.Interface, cfg *rest.Config,
	command []string, ptyHandler websocket.PtyHandler,
	namespace, podName, containerName string) error {

	req := k8sclient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}
	return executor.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
}

// HandleWebsocketLogs
func HandleWsLogs(w http.ResponseWriter, r *http.Request) {

}
