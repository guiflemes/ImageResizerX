package resizer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"nhooyr.io/websocket"
)

var hasWrite bool = false

type StubWsConn struct{}

func (ws *StubWsConn) CloseRead(ctx context.Context) context.Context {
	return ctx
}

func (ws *StubWsConn) Close(code websocket.StatusCode, reason string) error {
	return nil

}

func NewTestwebsocketClient(messageBuffer int) *websocketClient {
	return &websocketClient{
		subscriptions: make(map[*subscription]struct{}),
		messageBuffer: messageBuffer,
		writeTimeout: func(ctx context.Context, timeout time.Duration, conn WebsocketConn, msg Message) error {
			_, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			hasWrite = true
			return nil
		},
	}
}

func TestHandle(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	wsConn := &StubWsConn{}

	wsClient := NewTestwebsocketClient(10)

	sub := &subscription{
		message: make(chan Message, wsClient.messageBuffer),
	}
	wsClient.addSubscription(sub)
	defer wsClient.removeSubscription(sub)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(sub.message)
		for sub := range wsClient.subscriptions {
			sub.message <- Message{Action: "", DownloadUrl: ""}
		}
		wg.Done()
	}()

	err := wsClient.Handle(ctx, wsConn)
	assert.NoError(err)
	assert.True(hasWrite)

	hasWrite = false
	wg.Add(1)
	go func() {
		cancel()
		wg.Done()
	}()

	err = wsClient.Handle(ctx, wsConn)
	assert.Equal(err, context.Canceled)
	assert.False(hasWrite)

}

func TestBroadcast(t *testing.T) {
	assert := assert.New(t)

	wsClient := NewTestwebsocketClient(10)
	sub1 := &subscription{
		message: make(chan Message, wsClient.messageBuffer),
	}

	sub2 := &subscription{
		message: make(chan Message, wsClient.messageBuffer),
	}
	wsClient.addSubscription(sub1)
	wsClient.addSubscription(sub2)

	defer func() {
		wsClient.removeSubscription(sub1)
		wsClient.removeSubscription(sub2)
	}()

	msg := Message{Action: "", DownloadUrl: ""}
	wsClient.Brodcast(msg)

	assert.Equal(msg, <-sub1.message)
	assert.Equal(msg, <-sub2.message)
}
