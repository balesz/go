package queue

import (
	"context"
	"time"
)

// State is the type of the queue state holder
type State struct {
	ForceRun  bool      `firestore:"forceRun"`
	IsRunning bool      `firestore:"isRunning"`
	LastRun   time.Time `firestore:"lastRun, serverTimestamp"`
}

// Handler type define the interface of queue handling functions
type Handler interface {
	Handle(context.Context)
	NeedForceRun(context.Context)
}
