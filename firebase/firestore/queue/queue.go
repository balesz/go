package queue

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/balesz/go/firebase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const stateDocID = "$queue"

// IsStatePath returns true if the given path point to a queue state document
func IsStatePath(path string) bool {
	path = strings.TrimSpace(strings.TrimPrefix(path, "/"))
	if len(strings.Split(path, "/"))%2 != 0 {
		return false
	}
	return strings.HasSuffix(path, stateDocID)
}

// New creates a new queue
func New(path string) (queue Queue, err error) {
	if path == "" {
		err = fmt.Errorf("The path parameter is empty")
	} else if !regexp.MustCompile(`^/?\w+(?:/[\w\$]+)*$`).MatchString(path) {
		err = fmt.Errorf("The path parameter is invalid")
	}

	var trimmed = strings.TrimPrefix(strings.TrimSpace(path), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts)%2 == 0 {
		parts = parts[:len(parts)-1]
	}

	queue = Queue{
		rootPath:  strings.Join(parts, "/"),
		statePath: strings.Join(append(parts, stateDocID), "/"),
	}

	return
}

// Processor creates a new processor instance
func (queue Queue) Processor(workers ...Worker) (proc Processor, err error) {
	if queue.rootPath == "" || queue.statePath == "" {
		err = fmt.Errorf("Queue is not initialized")
	}
	proc = Processor{queue: queue, workers: workers}
	return
}

// Add method is add new worker to the processor
func (processor Processor) Add(worker Worker) {
	processor.workers = append(processor.workers, worker)
}

// Process method is execute the workers on the queue
func (processor Processor) Process(ctx context.Context, id string) error {
	firebase.InitializeClients()

	var process = processor.createProcess(ctx, id, 0)

	if err := process.start(); err != nil {
		return fmt.Errorf("process.start: %v", err)
	} else if err := process.handle(); err != nil {
		return fmt.Errorf("process.handle: %v", err)
	} else if err := process.stop(); err != nil {
		return fmt.Errorf("process.stop: %v", err)
	} else if err := process.forceRun(); err != nil {
		return fmt.Errorf("process.forceRun: %v", err)
	}

	return nil
}

func (processor Processor) createProcess(ctx context.Context, id string, idx int) process {
	return process{ctx: ctx, processID: id, processor: processor, worker: processor.workers[idx]}
}

func (process process) start() error {
	var (
		ctx         = process.ctx
		doc         = firebase.Firestore.Doc(process.processor.queue.statePath)
		maxAttempts = firestore.MaxAttempts(1)
		processID   = process.processID
	)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(doc)
		if err != nil && status.Code(err) != codes.NotFound {
			return fmt.Errorf("tran.Get: %v", err)
		}

		if !snap.Exists() {
			return tran.Create(doc, State{
				IsRunning: true,
				LastRunID: processID,
			})
		}

		var state State
		err = snap.DataTo(&state)
		if err != nil {
			return fmt.Errorf("snap.DataTo: %v", err)
		}

		if state.IsRunning {
			return fmt.Errorf("The queue runner is running")
		}

		return tran.Update(doc, []firestore.Update{
			{Path: "isRunning", Value: true},
			{Path: "lastRun", Value: firestore.ServerTimestamp},
			{Path: "lastRunID", Value: processID},
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}

func (process process) handle() error {
	var (
		ctx         = process.ctx
		doc         = firebase.Firestore.Doc(process.processor.queue.statePath)
		maxAttempts = firestore.MaxAttempts(5)
		processID   = process.processID
		worker      = process.worker
	)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(doc)
		if err != nil && status.Code(err) != codes.NotFound {
			return fmt.Errorf("tran.Get: %v", err)
		} else if !snap.Exists() {
			return fmt.Errorf("The queue state document not exists")
		}

		var state State
		err = snap.DataTo(&state)
		if err != nil {
			return fmt.Errorf("snap.DataTo: %v", err)
		} else if !state.IsRunning {
			return fmt.Errorf("The queue runner not running")
		} else if state.LastRunID != processID {
			return fmt.Errorf("The currently running is not this runner")
		}

		return worker.Execute(ctx, tran)
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}

func (process process) stop() error {
	var (
		ctx         = process.ctx
		doc         = firebase.Firestore.Doc(process.processor.queue.statePath)
		maxAttempts = firestore.MaxAttempts(5)
		processID   = process.processID
	)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(doc)
		if err != nil && status.Code(err) != codes.NotFound {
			return fmt.Errorf("tran.Get: %v", err)
		} else if !snap.Exists() {
			return fmt.Errorf("The queue state document not exists")
		}

		var state State
		err = snap.DataTo(&state)
		if err != nil {
			return fmt.Errorf("snap.DataTo: %v", err)
		} else if !state.IsRunning {
			return fmt.Errorf("The queue runner not running")
		} else if state.LastRunID != processID {
			return fmt.Errorf("The currently running is not this runner")
		}

		return tran.Update(doc, []firestore.Update{
			{Path: "isRunning", Value: false},
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}

func (process process) forceRun() error {
	var (
		ctx         = process.ctx
		doc         = firebase.Firestore.Doc(process.processor.queue.statePath)
		maxAttempts = firestore.MaxAttempts(2)
		processID   = process.processID
		worker      = process.worker
	)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		if !worker.NeedForceExec(ctx, tran) {
			return nil
		}

		snap, err := tran.Get(doc)
		if err != nil && status.Code(err) != codes.NotFound {
			return fmt.Errorf("tran.Get: %v", err)
		} else if !snap.Exists() {
			return fmt.Errorf("The queue state document not exists")
		}

		var state State
		err = snap.DataTo(&state)
		if err != nil {
			return fmt.Errorf("snap.DataTo: %v", err)
		} else if state.IsRunning {
			return fmt.Errorf("The queue runner is running")
		} else if state.LastRunID != processID {
			return fmt.Errorf("The currently running is not this runner")
		}

		return tran.Update(doc, []firestore.Update{
			{Path: "forceRun", Value: firestore.ServerTimestamp},
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}
