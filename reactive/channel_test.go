package reactive_test

import (
	"context"
	"testing"
	"time"

	"github.com/balesz/go/reactive"
)

func TestChannel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	timer := time.NewTicker(1 * time.Second)
	defer timer.Stop()

	channel := reactive.NewChannel(ctx)
	defer channel.Close()

	counter := 0

	for {
		select {
		case <-timer.C:
			counter += 1
			channel.Add(counter)
		case val, opened := <-channel.Get():
			if !opened {
				return
			}
			t.Log(val)
		}
	}
}
