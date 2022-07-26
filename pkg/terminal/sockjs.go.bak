package terminal

import (
	"encoding/json"
	"net/http"

	"github.com/forbearing/ratel-webterminal/pkg/errors"
	"github.com/forbearing/ratel-webterminal/pkg/k8s"
	session "github.com/forbearing/ratel-webterminal/pkg/terminal/sockjs"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	namespace     string
	podName       string
	containerName string

	terminalSessionList = session.SessionMap{Sessions: make(map[string]session.TerminalSession)}
)

// CreateAttachHandler is called from main for /api/sockjs
func CreateAttachHandler(path string) http.Handler {
	var (
		buf             string
		err             error
		msg             session.TerminalMessage
		terminalSession session.TerminalSession
	)

	// handleTerminalSession is called by net/http for any new /api/sockjs connections.
	handleTerminalSession := func(session sockjs.Session) {
		if buf, err = session.Recv(); err != nil {
			log.Errorf("handleTerminalSession: can't Recv: %v", err)
			return
		}
		if err = json.Unmarshal([]byte(buf), &msg); err != nil {
			log.Errorf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
			return
		}
		if msg.Op != "bind" {
			log.Errorf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
			return
		}
		if terminalSession = terminalSessionList.Get(msg.SessionID); terminalSession.ID == "" {
			log.Errorf("handleTerminalSession: can't find session '%s'", msg.SessionID)
			return
		}
		terminalSession.SockJSSession = session
		terminalSessionList.Set(msg.SessionID, terminalSession)
		terminalSession.Bound <- nil
	}
	return sockjs.NewHandler(path, sockjs.DefaultOptions, handleTerminalSession)
}

func HandleExecShell(ctx *gin.Context) {
	log.Info("Function: HandleExecShell")
	sessionID, err := session.GenTerminalSessionID()
	if err != nil {
		log.Errorf("session.GenTerminalSessionID error: ", err)
		errors.ResponseError(ctx, errors.CodeInternalError)
	}

	terminalSessionList.Set(sessionID, session.TerminalSession{
		ID:     sessionID,
		Bound:  make(chan error),
		SizeCh: make(chan remotecommand.TerminalSize),
	})
	log.Info("SessionsID: ", sessionID)
	go WaitForTerminal(ctx, k8s.Clientset(), k8s.RESTConfig(), sessionID)
}

// WaitForTerminal is called from  ratel-webterminal as a goroutine.
// Waits for the SockJS connection to be opened by the client,
// the session to be bound in handleTerminalSession.
func WaitForTerminal(ctx *gin.Context, k8sclient kubernetes.Interface, cfg *rest.Config, sessionID string) {
	parseParams(ctx)

	select {
	case <-terminalSessionList.Get(sessionID).Bound:
		log.Info("WaitForTerminal close session")
		close(terminalSessionList.Get(sessionID).Bound)

		var err error
		validShells := []string{"sh", "powershell", "cmd"}

		cmd := []string{"bash"}
		err = startProcess(ctx, k8sclient, cfg, cmd, terminalSessionList.Get(sessionID))
		if err != nil {
			for _, shell := range validShells {
				cmd := []string{shell}
				if err = startProcess(ctx, k8sclient, cfg, cmd, terminalSessionList.Get(sessionID)); err == nil {
					break
				}
			}
		}
		if err != nil {
			terminalSessionList.Close(sessionID, 2, err.Error())
			return
		}

		terminalSessionList.Close(sessionID, 1, "Process exited")
	}
}

func parseParams(c *gin.Context) {
	log.Info(namespace)
	log.Info(podName)
	log.Info(containerName)
	namespace = c.Param("namespace")
	if len(namespace) == 0 {
		errors.ResponseError(c, errors.CodeNamespaceNotSet)
	}
	podName = c.Param("pod")
	if len(podName) == 0 {
		errors.ResponseError(c, errors.CodePodNotSet)
	}
	containerName = c.Param("container")
	if len(containerName) == 0 {
		errors.ResponseError(c, errors.CodeContainerNotSet)
	}
}

// startProcess is cancelled by handleAttach
// Executed command in the container specified in request and connections it up
// with the ptyHandler (a session)
func startProcess(c *gin.Context, k8sclient kubernetes.Interface, cfg *rest.Config, command []string, ptyHandler session.PtyHandler) error {
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

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  ptyHandler,
		Stdout: ptyHandler,
		Stderr: ptyHandler,
		Tty:    true,
	})
	if err != nil {
		return err
	}

	return nil
}
