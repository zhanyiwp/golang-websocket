package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type CMap struct {
	m        sync.Mutex
	sessions map[string]*Session
}

func NewCMap() *CMap {
	s := make(map[string]*Session)
	return &CMap{sessions: s}
}
func (c *CMap) Add(uid string, s *Session) {
	c.m.Lock()
	defer c.m.Unlock()
	c.sessions[uid] = s
}

func (c *CMap) Delete(uid string) {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.sessions, uid)
}

type WebsocketService struct {
	onMessage    func(session *Session, Msg []byte)
	onConnect    func(*Session)
	onDisconnect func(*Session, error)
	sessions     *CMap
	hbInterval   time.Duration
	hbTimeout    time.Duration
	//stopCh     chan error
	addr     string
	uri      string
	upgrader websocket.Upgrader
}

func NewWebSocketService(addr, uri string) (*WebsocketService, error) {
	return &WebsocketService{addr: addr, uri: uri, upgrader: websocket.Upgrader{}, sessions: NewCMap()}, nil
}

func (w *WebsocketService) Server() {
	http.HandleFunc(w.uri, w.httpHandler)
	http.ListenAndServe(w.addr, nil)
}

func (w *WebsocketService) Stop() {
	//w.stopCh <- errors.New(reason)
}

func (w *WebsocketService) httpHandler(rw http.ResponseWriter, r *http.Request) {
	c, err := w.upgrader.Upgrade(rw, r, nil)
	if err != nil {
		//log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	w.connectHandler(c)
}

func (w *WebsocketService) connectHandler(c *websocket.Conn) {
	conn := NewConn(c, w.hbInterval, w.hbTimeout)
	session, err := NewSession(conn)
	if err != nil {
		conn.Close()
		return
	}
	// to do bind userid
	// 所有命令之前 必须有一个auth操作 绑定userid和连接  形成session
	w.sessions.Add(session.GetSessionID(), session)
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		conn.Close()
		w.sessions.Delete(session.GetSessionID())
	}()

	go conn.readCoroutine(ctx)
	go conn.writeCoroutine(ctx)

	if w.onConnect != nil {
		w.onConnect(session)
	}

	for {
		select {
		case err := <-conn.done:
			if w.onDisconnect != nil {
				w.onDisconnect(session, err)
			}
			return
		case msg := <-conn.messageCh:
			if w.onMessage != nil {
				w.onMessage(session, msg)
			}
		}
	}
}

func (w *WebsocketService) RegMessageHandler(handler func(*Session, []byte)) {
	w.onMessage = handler
}

// RegConnectHandler register connect handler
func (w *WebsocketService) RegConnectHandler(handler func(*Session)) {
	w.onConnect = handler
}

// RegDisconnectHandler register disconnect handler
func (w *WebsocketService) RegDisconnectHandler(handler func(*Session, error)) {
	w.onDisconnect = handler
}

func HandleMessage(s *Session, msg []byte) {
	fmt.Println("get msg ", string(msg))
	s.conn.SendMessage([]byte("hello world"))
}

func HandleConnect(s *Session) {
	fmt.Println("new connection ", s.conn.name)
}

func HandleDisconnect(s *Session, err error) {
	fmt.Println("dis connection  ", err)
}

func main() {
	addr := "localhost:8080"

	ws, err := NewWebSocketService(addr, "/echo")
	if err != nil {
		return
	}

	ws.RegMessageHandler(HandleMessage)
	ws.RegConnectHandler(HandleConnect)
	ws.RegDisconnectHandler(HandleDisconnect)

	ws.Server()
}
