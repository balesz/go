package queue

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

// Handler type define the interface of queue handling functions
type Handler interface {
	Handle(ctx context.Context, tran *firestore.Transaction) error
	NeedForceRun(ctx context.Context, tran *firestore.Transaction) bool
}

// Runner is the queue runner
type Runner struct {
	_documentPath string
	_initialized  bool
	ExecutionID   string
	Handler       Handler
	Path          string
}

// State is the type of the queue state holder
type State struct {
	ForceRun  time.Time `firestore:"forceRun,omitempty"`
	IsRunning bool      `firestore:"isRunning"`
	LastRun   time.Time `firestore:"lastRun,serverTimestamp"`
	LastRunID string    `firestore:"lastRunID"`
}
