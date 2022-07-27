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

// NewTerminalSession create TerminalSession
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

// remotecommand 包会启用 goroutine 调用 Next() 方法, 这个方法会一直阻塞来获取通道内容,

// sizeCh 包含了前端 TypeScript 发送过来的有关浏览器的长宽大小的信息, 一旦从 sizeCh
// 通道获取到了数据, remotecommand 包就会相应调整 pod 容器的 terminal 大小.

// 如果从 doneCh 获取到内容, 表明用户刷新了浏览器的 web终端或者其他 pod 容器的其他原因,
// 或者是网络原因, TerminalSession 将会关闭其内部 websocket, remotecommand 包
// 也会断开和 pod 容器建立的双向 shell streams 长连接.
func (t *TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeCh:
		return &size
	case <-t.doneCh:
		return nil
	}
}

// Read 从 TerminalSession 内部的 websocket 中读取数据, 并将读取到的输出作为容器的 stdin

// 1.remotecommand.NewSPDYExecutor() 函数创建一个 Executor 来和 pod 容器建立一个
//   多路复用的双向 shell streams 连接.
// 2.remotecommand 包通过 Executor.Stream() 选项来设置 pod 容器的 stdin, stdout, stderr
//   这里传入的是一个 TerminalSession 对象,
// 3.也就是说, pod 容器的 stdin 是来自 TerminalSession 内部的 websocket.ReadMessage()
//   pod 容器 stdout, stderr 会通过调用 TerminalSession 内部的 websocket 的 WriteMessage()
//   方法将 pod 容器的 stdout, stder 内容写入到 TerminalSession 内部的 websocket 中

// 1.前端 TypeScript 代码通过 conn.send() 函数将用户在浏览器 web 终端输出的 shell 命令
//   发送到 TerminalSession 的 websocket 连接中, 并将 Op 标志设置为 "stdin"
// 2.检查 Op 标志, 如果为 "stdin", 将 TerminalSession 内部 websocket 中的内容
//   作为 pod 容器的 stdin 发送给 pod 容器, 来执行用户输入的 shell 命令
//   具体代码见 ./frontend/terminal.js 34,35,46,47 行.
// 3.检查 Op 标志, 如果为 "resize", remotecommand 则会根据前端 TypeScript 代码传来的
//   浏览器长宽大小来调整容器的 Terminal 大小.
//   具体代码将 ./frontend/terminal.js 39,40 行.
func (t *TerminalSession) Read(p []byte) (int, error) {
	_, message, err := t.conn.ReadMessage()
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			log.Println("closed network connection")
			return copy(p, END_OF_TRANSMISSION), nil
		} else {
			log.Printf("read message err: %v", err)
			return copy(p, END_OF_TRANSMISSION), err
		}
		//return copy(p, EndOfTransmission), err
	}
	var msg TerminalMessage
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		log.Printf("read parse message err: %v", err)
		return copy(p, END_OF_TRANSMISSION), err
	}
	switch msg.Op {
	case "stdin":
		return copy(p, msg.Data), nil
	case "resize":
		t.sizeCh <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		log.Printf("unknown message type '%s'", msg.Op)
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// Write 将 pod 容器的任何 stdout 和 stderr 输出写入到 TerminalSession 内部的 websocket 中,
// 前端 TypeScript 代码从 TerminalSession 内部的 websocket 读取数据并输出到 浏览器的 web terminal 中.

// 1.remotecommand 通过 Executor.Stream() 方法设置好了 pod 容器的 stdout, stderr.
// 2.remotecommand 可以获取 pod 容器的任何 stdout, stderr 输出, 并将 stdout, stderr
//   输出内容写入到 TerminalSession 内部的 websocket 中. 并将 Op 标志设置为 "stdout".
// 3.前端 TypeScript 代码检查 Op 标志是否为 "stdout", 如果是 "stdout", 则将通过
//   term.write(msg.data) 函数将 TerminalSession 内部 websocket 的内容输出到浏览器
//   的 web 终端上.
//
// 具体代码见 ./frontend/terminal.js 的 52,53 行.
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
