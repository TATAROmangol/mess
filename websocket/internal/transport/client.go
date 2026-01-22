package transport

import (
	"fmt"
	"time"

	wsdto "github.com/TATAROmangol/mess/shared/dto/ws"
	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
)

type Client struct {
	SubjectID string
	Send      chan *wsdto.WSMessage
	cfg       ClientConfig
	hub       *Hub
	conn      *websocket.Conn
}

func NewClient(subjectID string, conn *websocket.Conn, cfg ClientConfig, hub *Hub) *Client {
	return &Client{
		SubjectID: subjectID,
		Send:      make(chan *wsdto.WSMessage, cfg.MessageBuffer),
		cfg:       cfg,
		hub:       hub,
		conn:      conn,
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout))
		return nil
	})

	for {
		mt, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.sendError(err)
			}
			break
		}

		if mt == websocket.CloseMessage {
			break
		}

		_ = message
	}
}

func (c *Client) sendError(err error) {
	c.hub.lg.Error(fmt.Errorf("subj: %v, err: %w", c.SubjectID, err))
}

func (c *Client) writePump() {
	ticker := time.NewTicker(c.cfg.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteTimeout))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.sendError(err)
				return
			}
			msg, err := message.GetBytes()
			if err != nil {
				c.sendError(err)
				return
			}

			w.Write(msg)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				msg, err := (<-c.Send).GetBytes()
				if err != nil {
					c.sendError(err)
					return
				}
				w.Write(msg)
			}

			if err := w.Close(); err != nil {
				c.sendError(err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.sendError(err)
				return
			}
		}
	}
}
