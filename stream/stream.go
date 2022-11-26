package stream

import "context"

type Stream interface {
	Chan() <-chan interface{}
	Add(interface{})
	Close()
	IsClosed() bool
}

func NewStream(ctx context.Context) Stream {
	strm := stream{make(chan interface{})}
	go strm.closeWhenDone(ctx)
	return &strm
}

type stream struct {
	data chan interface{}
}

func (st *stream) Chan() <-chan interface{} {
	return st.data
}

func (st *stream) Add(val interface{}) {
	if !st.IsClosed() {
		go func() { st.data <- val }()
	}
}

func (st *stream) Close() {
	if !st.IsClosed() {
		close(st.data)
	}
}

func (st *stream) IsClosed() bool {
	select {
	case _, open := <-st.data:
		return !open
	default:
		return false
	}
}

func (st *stream) closeWhenDone(ctx context.Context) {
	if ctx.Done() == nil {
		return
	}
	defer st.Close()
	<-ctx.Done()
}
