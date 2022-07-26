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

// 前端为 xterm.js, 前端代码的主要函数功能
//    term.write("message"):   将 message 信息写到浏览器上
//    conn.send():             向 websocket 中写数据

// HandleWsTerminal

// 这里主要有三个数据结构: TerminalMessage, TerminalSession, PtyHandler
// TerminalMessage 相当于一个协议, 支持支持3种, stdin, stdout, resize

// TerminalSession 包含三个字段: conn, sizeCh, doneCh
//     conn 就是一个 websocket 连接
//     sizeCh 是一个 remotecommand.TerminalSize 类型, 该类型包含了浏览器的长宽信息
//     doneCh 是一ge channel 类型, 用来标记 TerminalSession 是否关闭了.

// TerminalSession 对象有4个方法, Read(), Write(), Next(), Close().
//     Read() 用来从 TerminalSession 的 websocket 读取数据, 实现了 io.Reader 接口
//     Write() 用来从 TerminalSession 的 websocket 写数据, 实现了 io.Writer 接口
//     Next() 实现了 remotecommand.TerminalSizeQueue 接口
// PtyHandler 是一个接口, 其包含了三个子接口, 分别是: io.Reader, io.Writer, remotecommand.TerminalSizeQueue
// 所以 TerminalSession 对象实现了 PtyHandler 接口.

// 每一个浏览器终端都会使用一个 TerminalSession

// podHandler.Execute(podName, containerName, []string{"bash"}, pty) 会调用以下这段代码.
//exec, err := remotecommand.NewSPDYExecutor(h.config, "POST", req.URL())
//if err != nil {
//    return err
//}
//return exec.Stream(remotecommand.StreamOptions{
//    Stdin:             pty,
//    Stdout:            pty,
//    Stderr:            pty,
//    TerminalSizeQueue: pty,
//    Tty:               true,
//})
// 如何理解这段代码:
// 1.remotecommand.NewSPDYExecutor 会与 pod 容器建立连接,并将连接升级到一个多路复用流连接.
// 2.exec.Stream(remotecommand.StreamOptions{}) 会建立一个标准的 shell streams, 并设置
//   容器的 stdin, stdout, stderr
//   这里传入的是一个 PtyHandler 接口的实例, 其实是一个 TerminalSession 的对象.
//   因为 TerminalSession 实现了 PtyHandler 接口(上面已经讲了).
//   remotecommand 会调用 TerminalSession 的 Write 方法将 pod 容器的任何输出内容,
//   将通过 TerminalSession 的 Write 方法写入到 TerminalSession 的 websocket 连接中
//   然后前端 TypeScript 代码会读取 websocket, 将 pod 容器的输出反映到浏览器 web terminal 上.

// pod 容器输出:
// 1.remotecommand.NewSPDYExecutor() 函数创建一个 Executor, 将会和 pod 容器连理一个多路复用的双向流连接.
// 2.remotecommand 调用 Executor.Stream 方法来初始化 pod 容器的 stdin, stdout, stderr.
//   这里是用一个 TerminalSession 对象来初始化 pod 容器的 stdin, stdout, stderr.
// 3.pod 容器的任何输出, 都会被 remotecommand 调用 TerminalSession 对象的 Write 方法
//   写入到 TerminalSession 内部的一个 websocket 连接中.
// 4.前端 TypeScript 代码会 websocket 读取内容并将这些内容反映到浏览器的 web terminal 上.

// pod 容器获取用户输出的命令
// 1.前端 TypeScript 代码会将用户在浏览器输入的或复制的 shell 命令写入 TerminalSession
//   的 websocket 连接中, remotecommand 调用 TerminalSession 的 Read 方法将
//

// 具体流程
// 1.前端 TypeScript 代码会将用户在浏览器输入的或复制的 shell 命令, 写入到 TerminalSession 中
//   具体代码为 ./frontend/terminal.js 的第 34,35 行代码
//   协议为 TerminalMessage, 其中 Data 包含了具体的传输内容, Op 是一个标志, 这里为 stdin
//   表示向 TerminalSession 写内容,
//   向 Stream 中写数据, TerminalMessage 的 Data 包含了具体的内容, Op 是一个标志

// 1.TypeScript 代码将用户输出的或复制的 shell 命令写入到 remotecommand 建立的
//   长连接中, Op 为 "stdin". remotecommand 会调用 TerminalSession 的 Read 方法
//   将 shell 命令传给容器执行.
//   具体 TypeScript 代码为 ./frontend/terminal.js 的 34,35,46,47 行.
// 2.remotecommand 包会将容器的 shell 命令输出信息, 通过调用 TerminalSession 的
//   Write 方法写给前端 TypeScript 并附带 Op 为 stdout 的标志. 前端 TypeScript 代码
//   检查 Op 标志, 如果 Op 标志为 "stdout", 则将容器的输出写入到浏览器.
//   具体 TypeScript 代码为 ./frontend/terminal.js 的 52,53 行.
// 3.浏览器的尺寸发生变化, TypeScript 代码会将浏览器的持续信息写入到 remotecommand
//   建立的长连接中,并且将 Op 标志设置为 "resize", remotecommand 就会调整容器
//   terminal 的大小.
//   具体 TypeScript 代码为 ./frontend/terminal.js 的 38,39 行代码.

func HandleWsTerminal(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container"]
	log.Infof("exec pod: %s/%s, container: %s", namespace, podName, containerName)

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
// 前端 TypeScript 代码会通过这个 uri 合成一个新的 uri, 然后前端的 TypeScript 脚本会调用
// 我们的 RESTful API, 也就是 /ws/{namespace}/{pod}/{container}/logs

// 前端 TypeScript 代码关键行在 frontend/logs.js 的 25行 和 45行.
// 25行处理 http://localhost:8080/logs?namespace=default&pod=nginx&container=nginx 请求生成一个新的 uri
// 45行根据新的 uri 调用 ratel-webterminal 的 RESTful API, 也就是 /ws/{namespace}/{pod}/{container}/logs

// 总结:
// 1.前端的 TypeScript 代码会转换浏览器请求的 uri
//   比如 http://localhost:8080/logs?namespace=default&pod=nginx&container=nginx
// 2.前端的 TypeScript 代码再调用 ratel-webtermal 的 api,
//   也就是这里的 /ws/{namespace}/{pod}/{container}/logs
func HandleWsLogs(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container"]
	tailLines, _ := strconv.ParseInt(r.URL.Query().Get("tail"), 10, 64)
	log.Infof("get pod logs: %s/%s, container: %s, tailLines: %d\n", namespace, podName, containerName, tailLines)

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
	// 2.前端的 TypeScript 脚本会读取 websocket 中的 pod 日志.
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
