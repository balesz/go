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

type iqueue interface {
	Processor(workers ...Worker) Processor
}

// Processor is the struct of the queue processor
type Processor struct {
	queue   Queue
	workers []Worker
}

type iprocessor interface {
	Add(worker Worker)
	Process(ctx context.Context, id string) error
	createProcess(ctx context.Context, id string, worker Worker) process
}

type process struct {
	ctx       context.Context
	processID string
	processor Processor
	worker    Worker
}

type iprocess interface {
	start() error
	handle() error
	stop() error
	forceRun() error
}

// Worker defines the queue worker interface
type Worker interface {
	Execute(ctx context.Context, tran *firestore.Transaction) error
	NeedForceExec(ctx context.Context, tran *firestore.Transaction) bool
}

// State is the type of the queue state holder
type State struct {
	Disabled    bool                   `firestore:"disabled"`
	ForceRunRef *firestore.DocumentRef `firestore:"forceRunRef"`
	IsRunning   bool                   `firestore:"isRunning"`
	LastRun     time.Time              `firestore:"lastRun,serverTimestamp"`
	LastRunID   string                 `firestore:"lastRunID"`
}

// ForceRunState is the type of the force run document
type ForceRunState struct {
	QueueStateRef *firestore.DocumentRef `firestore:"queueStateRef"`
	Trigger       time.Time              `firestore:"trigger,serverTimestamp"`
}
