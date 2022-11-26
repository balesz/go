package stream_test

import (
	"context"
	"testing"
	"time"

	"github.com/balesz/go/stream"
)

func TestMain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	timer := time.NewTicker(1 * time.Second)
	defer timer.Stop()

	strm := stream.NewStream(ctx)
	defer strm.Close()

	counter := 0

	for {
		select {
		case <-timer.C:
			counter += 1
			strm.Add(counter)
		case val, opened := <-strm.Chan():
			if !opened {
				return
			}
			t.Log(val)
		default:
		}
	}
}
