package reactive_test

import (
	"testing"
	"time"
)

func TestTest(t *testing.T) {
	ch := make(chan int)

	var listeners = map[int]func(int){}

	addLlistener := func(i int) {
		listeners[i] = func(val int) {
			t.Logf("Listener %v: %v", i, val)
		}
	}

	removeListener := func(i int) {
		delete(listeners, i)
	}

	go func() {
		for {
			if val, more := <-ch; more {
				for _, fn := range listeners {
					fn(val)
				}
			} else {
				return
			}
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)
		removeListener(2)
	}()

	for i := range []int{1, 2, 3} {
		addLlistener(i + 1)
	}

	timeout := time.After(10 * time.Second)

	counter := 0
	for {
		select {
		case <-time.After(time.Second):
			counter += 1
			ch <- counter
		case <-timeout:
			t.Log("Close")
			close(ch)
		case _, more := <-ch:
			if !more {
				return
			}
		}
	}
}
