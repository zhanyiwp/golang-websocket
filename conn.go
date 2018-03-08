package main

import (
	"context"

	"time"

	"github.com/gorilla/websocket"
)

// Conn wrap websocket.Conn
type Conn struct {
	sid        string
	rawConn    *websocket.Conn
	sendCh     chan []byte
	done       chan error
	hbTimer    *time.Timer
	name       string
	messageCh  chan []byte
	hbInterval time.Duration
	hbTimeout  time.Duration
}

// GetName Get conn name
func (c *Conn) GetName() string {
	return c.name
}

// NewConn create new conn
func NewConn(c *websocket.Conn, hbInterval time.Duration, hbTimeout time.Duration) *Conn {
	conn := &Conn{
		rawConn:    c,
		sendCh:     make(chan []byte, 100),
		done:       make(chan error),
		messageCh:  make(chan []byte, 100),
		hbInterval: hbInterval,
		hbTimeout:  hbTimeout,
	}

	conn.name = c.RemoteAddr().String()
	conn.hbTimer = time.NewTimer(conn.hbInterval)

	if conn.hbInterval == 0 {
		conn.hbTimer.Stop()
	}

	return conn
}

// Close close connection
func (c *Conn) Close() {
	c.hbTimer.Stop()
	c.rawConn.Close()
}

// SendMessage send message
func (c *Conn) SendMessage(msg []byte) error {
	c.sendCh <- msg
	return nil
}

// writeCoroutine write coroutine
func (c *Conn) writeCoroutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case pkt := <-c.sendCh:

			if pkt == nil {
				continue
			}

			if err := c.rawConn.WriteMessage(websocket.BinaryMessage, pkt); err != nil {
				c.done <- err
			}

		case <-c.hbTimer.C:
			// to do MsgHeartbeat
			//hbMessage := NewMessage(MsgHeartbeat, hbData)
			//c.SendMessage(hbMessage)
		}
	}
}

// readCoroutine read coroutine
func (c *Conn) readCoroutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			// 设置超时
			if c.hbInterval > 0 {
				err := c.rawConn.SetReadDeadline(time.Now().Add(c.hbTimeout))
				if err != nil {
					c.done <- err
					continue
				}
			}
			// 读取长度
			_, msg, err := c.rawConn.ReadMessage()
			if err != nil {
				c.done <- err
				continue
			}
			// 设置心跳timer
			if c.hbInterval > 0 {
				c.hbTimer.Reset(c.hbInterval)
			}
			c.messageCh <- msg
		}
	}
}
