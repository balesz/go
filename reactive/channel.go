package reactive

import "context"

func NewChannel(ctx context.Context) *channel {
	ch := channel{make(chan interface{})}
	go ch.closeWhenDone(ctx)
	return &ch
}

type channel struct {
	data chan interface{}
}

func (ch *channel) Get() <-chan interface{} {
	return ch.data
}

func (ch *channel) Add(val interface{}) {
	if !ch.IsClosed() {
		go func() { ch.data <- val }()
	}
}

func (ch *channel) Close() {
	if !ch.IsClosed() {
		close(ch.data)
	}
}

func (ch *channel) IsClosed() bool {
	select {
	case _, open := <-ch.data:
		return !open
	default:
		return false
	}
}

func (ch *channel) closeWhenDone(ctx context.Context) {
	if ctx.Done() == nil {
		return
	}
	defer ch.Close()
	<-ctx.Done()
}
