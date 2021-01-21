package queue

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

// Queue is the struct of the queue
type Queue struct {
	forceRunPath string
	statePath    string
}

// Task is the struct of the queue processor
type Task struct {
	ID     string
	queue  Queue
	worker Worker
}

// Worker defines the queue worker interface
type Worker interface {
	Execute(ctx context.Context, tran *firestore.Transaction) error
	NeedForceExec(ctx context.Context, tran *firestore.Transaction) error
}

// State is the type of the queue state holder
type State struct {
	Disabled    bool                   `firestore:"disabled"`
	ForceRunRef *firestore.DocumentRef `firestore:"forceRunRef"`
	IsRunning   bool                   `firestore:"isRunning"`
	LastRun     time.Time              `firestore:"lastRun,serverTimestamp"`
	LastTaskID  string                 `firestore:"lastTaskID"`
}

// ForceRunState is the type of the force run document
type ForceRunState struct {
	QueueStateRef *firestore.DocumentRef `firestore:"queueStateRef"`
	Trigger       time.Time              `firestore:"trigger,serverTimestamp"`
}
