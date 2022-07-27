package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/remotecommand"
)

// NewTerminalSessionWs create TerminalSession
func NewTerminalSessionWs(conn *websocket.Conn) *TerminalSession {
	return &TerminalSession{
		conn:   conn,
		sizeCh: make(chan remotecommand.TerminalSize),
		doneCh: make(chan struct{}),
	}
}

// NewTerminalSession 创建一个 TerminalSession 对象,  同时将 http 连接升级到 websocket 并放入 TerminalSession 对象.
// 后续前端 JavaScript 代码可以向 TerminalSession 内部维护的 websocket 写数据和读取数据
func NewTerminalSession(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*TerminalSession, error) {
	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	session := &TerminalSession{
		conn:   conn,
		sizeCh: make(chan remotecommand.TerminalSize),
		doneCh: make(chan struct{}),
	}
	return session, nil
}

// remotecommand 会循环调用 Next() 方法
func (t *TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	// sizeCh 包含了前端 TypeScript 发送过来的有关浏览器的长宽大小的信息, 一旦从 sizeCh
	// 通道获取到了数据, remotecommand 包就会相应调整 pod 容器的 terminal 大小.
	case size := <-t.sizeCh:
		return &size

		// 如果从 doneCh 被关闭了, 表明用户刷新了浏览器的 web 终端或者其他 pod 容器的其他原因,
		// 或者是网络原因, TerminalSession 将会关闭其内部 websocket, remotecommand 包
		// 也会断开和 pod 容器建立的双向 shell streams 长连接.
	case <-t.doneCh:
		return nil
	}
}

// remotecommand.NewSPDYExecutor() 函数创建一个 Executor 来和 pod 容器建立一个双向的 shell streams 连接.
// 再通过 Executor.Stream() 方法设置好了 pod 容器的 stdin, stdout, stderr, 最终效果是:
// 1. pod 容器的任何输入都来自 TerminalSession 内部维护的 websocket.
// 2. pod 容器的任何输出都会写入 TerminalSession 内部维护的 websocket.

// 1.Read 方法用来从 TerminalSession 内部维护的 websocket 读取数据.
// 2.remotecommand 包会调用 TerminalSession 的 Read 方法来获取数据, 作为 pod 容器的 stdin
// 3.前端 JavaScript 代码将用户在浏览器 web 终端上输入的 shell 指令写入到 TerminalSession 内部的 websocket.
// 4.最终 pod 容器知道要执行哪个命令.
func (t *TerminalSession) Read(p []byte) (int, error) {
	// 调用 websocket.ReadMessage() 方法从 websocket 读取数据, 数据类型主要有两种
	// 一种是用户输入的 shell 指令, 一种是浏览器长宽大小信息.
	_, message, err := t.conn.ReadMessage()
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			log.Println("closed network connection")
			return copy(p, END_OF_TRANSMISSION), nil
		} else {
			log.Printf("read message err: %v", err)
			return copy(p, END_OF_TRANSMISSION), err
		}
	}
	var msg TerminalMessage
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		log.Printf("read parse message err: %v", err)
		return copy(p, END_OF_TRANSMISSION), err
	}

	// 如果 Op 标志位为 stdin, 表示是用户输入的 shell 指令
	// 具体前端 JavaScript 代码为 ./frontend/terminal.js 34, 35 行
	switch msg.Op {
	case "stdin":
		return copy(p, msg.Data), nil

	// 如果 Op 标志位为 resize, 表示是浏览器长宽大小信息.
	// 则向 sizeCh 通道发送当前浏览器长宽大小信息, remotecommand 包会循环调用 TerminalSession 的 Next() 方法
	// 在该 Next() 方法中, remotecommand 包接收到了浏览器新的长宽大小, 就会调整 pod 容器的终端大小.
	// 具体前端 JavaScript 代码为 "./frontend/terminal.js" 的 39,40行
	case "resize":
		t.sizeCh <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		log.Printf("unknown message type '%s'", msg.Op)
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// remotecommand.NewSPDYExecutor() 函数创建一个 Executor 来和 pod 容器建立一个双向的 shell streams 连接.
// 再通过 Executor.Stream() 方法设置好了 pod 容器的 stdin, stdout, stderr, 最终效果是:
// 1.pod 容器的任何输入都来自 TerminalSession 内部维护的 websocket.
// 2.pod 容器的任何输出都会写入 TerminalSession 内部维护的 websocket.

// 1.Write 方法用来将数据写入到 TerminalSession 内部维护的 websocket.
// 2.remotecommand 包会调用 TerminalSession 的 Write 方法将 pod 容器的任何 stdout, stderr 输出写入到 TerminalSession
//   内部维护的 websocket
// 3.前端 JavaScript 代码从 TerminalSession 内部维护的 websocket 读取数据并输出到浏览器的 web terminal 上.
// 4.最终用户在 web 终端上得到自己命令的输出结果.
func (t *TerminalSession) Write(p []byte) (int, error) {
	msg, err := json.Marshal(TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	})
	if err != nil {
		log.Printf("write parse message err: %v", err)
		return 0, err
	}
	if err := t.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Printf("write message err: %v", err)
		return 0, err
	}
	return len(p), nil
}

// 浏览器 web 终端重新刷新了, 集群 pod 容器故障, 或者其他网络原因等, 将会关闭一个 TerminalSession

// Close 函数将会关闭 TerminalSession 内部的 websocket, 并关闭 doneCh 通道.
// remotecommand 包调用的 Next() 函数感知到 doneCh 通道关闭了,也就会关闭与 pod 容器
// 建立的双向 shell streams 长连接.
func (t *TerminalSession) Close() error {
	close(t.doneCh)
	return t.conn.Close()
}
