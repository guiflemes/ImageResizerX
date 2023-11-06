package resizer

import (
	"context"
	"imageResizerX/logs"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type WebsocketConn interface {
	CloseRead(context.Context) context.Context
	Close(code websocket.StatusCode, reason string) error
}

type Message struct {
	Action      string `json:"action"`
	DownloadUrl string `json:"download_url"`
}

type subscription struct {
	message   chan Message
	closeSlow func()
}

type websocketClient struct {
	subscriptions map[*subscription]struct{}
	lock          sync.RWMutex
	messageBuffer int
	writeTimeout  func(ctx context.Context, timeout time.Duration, conn WebsocketConn, msg Message) error
}

func DefaultwebsocketClient() *websocketClient {
	return &websocketClient{
		subscriptions: make(map[*subscription]struct{}),
		messageBuffer: 16,
		writeTimeout: func(ctx context.Context, timeout time.Duration, conn WebsocketConn, msg Message) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return wsjson.Write(ctx, conn.(*websocket.Conn), msg)
		},
	}
}

func (c *websocketClient) Handle(ctx context.Context, conn WebsocketConn) error {
	ctx = conn.CloseRead(ctx)

	s := &subscription{
		message: make(chan Message, c.messageBuffer),
		closeSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		},
	}

	c.addSubscription(s)
	defer c.removeSubscription(s)

	for {
		select {
		case msg := <-s.message:
			c.writeTimeout(ctx, time.Second*5, conn, msg)
			return nil

		case <-ctx.Done():
			logs.Logger.Info("close websocket connection")
			return ctx.Err()
		}

	}
}

func (c *websocketClient) addSubscription(s *subscription) {
	c.lock.Lock()
	c.subscriptions[s] = struct{}{}
	c.lock.Unlock()
	logs.Logger.Info("add subscription")
}

func (c *websocketClient) removeSubscription(s *subscription) {
	c.lock.Lock()
	delete(c.subscriptions, s)
	c.lock.Unlock()
	logs.Logger.Info("remove Subscription")
}

func (c *websocketClient) SubscriptionCount() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.subscriptions)
}

func (c *websocketClient) Brodcast(msg Message) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for s := range c.subscriptions {
		select {
		case s.message <- msg:
		default:
			go s.closeSlow()
		}
	}

	logs.Logger.Info("Brodcast message to all subscriptions")
}
