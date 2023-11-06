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
	go func() {
		defer close(sub.message)
		for sub := range wsClient.subscriptions {
			sub.message <- Message{Action: "", DownloadUrl: ""}
		}
		wg.Done()
	}()
	wg.Add(1)

	err := wsClient.Handle(ctx, wsConn)
	assert.NoError(err)
	assert.True(hasWrite)

	hasWrite = false
	go func() {
		cancel()
		wg.Done()
	}()
	wg.Add(1)

	err = wsClient.Handle(ctx, wsConn)
	assert.Equal(err, context.Canceled)
	assert.False(hasWrite)

}
