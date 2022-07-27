package websocket

import (
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
)

const END_OF_TRANSMISSION = "\u0004"

// PtyHandler 是一个接口包含了三个子接口: io.Reader, io.Writer, remotecommand.TerminalSizeQueue
// io.Reader 接口有一个 Read() 方法
// io.Writer 接口有一个 Write() 方法
// remotecommand.TerminalSizeQueue 接口有一个 Next() 方法
// TerminalSession 对象有: Read(), Write(), Next() 方法, 所有 TerminalSession 对象实现了 PtyHandler 接口

// remotecommand 包的 NewSPDYExecutor() 函数会创建一个 Executor 对象来和 pod 容器建立双向的 shell streams 长连接.
// 再通过 Executor.Stream() 方法用一个 PtyHandler 实例来设置 pod 容器的 stdin, stdout, stderr. 这个 PtyHandler 实例
// 也就是 TerminalSession 对象.

// 当 remotecommand 与 pod 容器建立好双向的 shell streams 长连接, 并用一个 TerminalSession 设置好 pod 容器的 stdin,stdout,stderr 之后
// 1. pod 容器的 stdin(), 来自 TerminalSession.Read() 方法. Read() 方法会从其内部的 websocket 中读取数据, 也就是浏览器
//    web 终端中用户输入的 shell 指令.
// 2. pod 容器的 stdout, stderr 会通过 TerminalSession.Write() 方法写入到其内部的 websocket. websocket 会被前端 JavaScript
//    代码读取并输出到浏览器 web 终端上
// 3. 最终用户看到了自己 shell 指令的输出结果.
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

// TerminalSession 三个字段/属性.
// conn:    内部维护的一个 websocket 连接, NewTerminalSession() 函数可以把 http 连接升级到 websocket.
//          升级得到的 websocket 就存放在这里.
//          后续浏览器 JavaScript 会将用户的 shell 指令写入到该 websocket, 并调用 TerminalSession 的 Read() 方法写入到 pod 容器
//          pod 容器的输出会调用 TerminalSession 的 Write() 方法写入到该 websocket, 前端 JavaScript 代码就可以从该 websocket
//          读取容器的输出内容,并写到 web 终端浏览器上
// sizeCh:  是一个 remotecommand.TerminalSize, 代表浏览器 web 终端的长宽大小
// doneCh:  当 remotecommand 包与 pod 容器建立的双向 shell streams 长连接断开后(比如用户刷新浏览器等操作导致的),
//          remotecommand 包会向 doneCh 发送一个空数据, TerminalSession 会关闭内部维护的 websocket
type TerminalSession struct {
	conn   *websocket.Conn
	sizeCh chan remotecommand.TerminalSize
	doneCh chan struct{}
}

// TerminalMessage 是前端 JavaScript 代码和 TerminalSession 内部维护的 websocket 之间的通信协议.

// Op:     标志位,用来标记通信数据的类型.
//         如果为 stdin,  表示前端 JavaScript 代码将用户输出的 shell 指令发送到 TerminalSession 内部维护的 websocket.
//         如果为 stdout, 表示前端 JavaScript 代码将从 TerminalSession 内部维护的 websocket 读取数据并输出到浏览器 web 终端上.
//         如果为 resize, 表示前端 JavaScript 代码将浏览器到长宽大小信息发送到 TerminalSession 内部维护 websocket.
// Data:   前端 JavaScript 代码从 TerminalSession 内部内部维护的 websocket 中写入或读取的数据, Op 为 stdin 或 stdout
// Rows,Cols:  浏览器的长宽大小信息, Op 为 resize.
type TerminalMessage struct {
	Op   string `json:"op"`
	Data string `json:"data"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

var upgrader = func() websocket.Upgrader {
	upgrader := websocket.Upgrader{}
	upgrader.HandshakeTimeout = time.Second * 2
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	return upgrader
}()

type Logger struct {
	conn *websocket.Conn
}
