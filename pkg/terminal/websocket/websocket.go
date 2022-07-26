package websocket

import (
	"context"
	"net/http"
	"strconv"

	"github.com/forbearing/k8s/pod"
	"github.com/forbearing/ratel-webterminal/pkg/args"
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

// HandleWsTerminal
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

	podHandler, err := pod.New(context.TODO(), args.GetKubeConfigFile(), namespace)
	err = podHandler.Execute(podName, containerName, []string{"bash"}, pty)
	if err != nil {
		if err = podHandler.Execute(podName, containerName, []string{"sh"}, pty); err != nil {
			log.Error("create pod shell error: ", err)
		}
	}
}

// HandleWsLogs 用来处理 api 为 /ws/{amespace}/{pod}/{container}/logs 的请求

// 假设 namespace=default, pod=nginx, container=nginx, 监听在 0.0.0.0:8080
// 实际在浏览器中访问的路径为: http://localhost:8080/logs?namespace=default&pod=nginx&container=nginx
// 注意这里的 uri 和我们实际的 api 不同. 为什么呢?
// 因为当用浏览器访问 http://localhost:8080/logs?namespace=default&pod=nginx&container=nginx 时,
// 前端 javascript 代码会通过这个 uri 合成一个新的 uri, 然后前端的 javascript 脚本会调用
// 我们的 RESTful API, 也就是 /ws/{namespace}/{pod}/{container}/logs

// 前端 javascript 代码关键行在 frontend/logs.js 的 25行 和 45行.
// 25行处理 http://localhost:8080/logs?namespace=default&pod=nginx&container=nginx 请求生成一个新的 uri
// 45行根据新的 uri 调用 ratel-webterminal 的 RESTful API, 也就是 /ws/{namespace}/{pod}/{container}/logs

// 总结:
// 1.前端的 javascript 代码会转换浏览器请求的 uri
//   比如 http://localhost:8080/logs?namespace=default&pod=nginx&container=nginx
// 2.前端的 javascript 代码再调用 ratel-webtermal 的 api,
//   也就是这里的 /ws/{namespace}/{pod}/{container}/logs
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
		log.Println("close logs session.")
		writer.Close()
	}()

	// TailLines 字段用来指定获取多少行 Pod 的日志
	// 如果没有设置 TailLines, 就可以查看到 pod 中所有的日志.

	// Follow 字段的功能类似于 kubectl logs 命令加了一个 -f 标志, 用来持续追踪 pod
	// 接下来产生的日志.
	// 如果不将 Follow 设置为 true, 就无法动态获取 pod 接下来生成的日志.
	// 如果想通过浏览器来持续观察一个 pod 的日志, Follow 应该总是设置成 True.

	// 这个 writer 有一个 write 方法, 实现了 io.Writer 接口. 调用这个 writer 的
	// write 方法, 就会执行 conn.WriteMessage 函数, 即向 websocket 写数据.
	// 总流程为:
	// 1.podHandler.Log() 指定 pod 的日志, 默认日志是写入到标准输出, 但是我们通过
	//   logOptions 设置了 Writer 字段为我们指定的 writer. 也就是说会将 pod 的日志
	//   远远不断的写入到 writer 对象封装的 websocket 连接中.
	// 2.前端的 javascript 脚本会读取 websocket 中的 pod 日志.
	//   然后我们就可以在浏览器中查看到这个 pod 的日志.

	var logOptions pod.LogOptions
	logOptions.Writer = writer
	logOptions.Follow = true
	if tailLines != 0 {
		logOptions.TailLines = &tailLines
	}
	podHandler, err := pod.New(context.TODO(), args.GetKubeConfigFile(), namespace)
	if err != nil {
		log.Error("get pod handler error")
		return
	}
	if err = podHandler.Log(podName, &logOptions); err != nil {
		log.Error("get pod log error")
	}
}
