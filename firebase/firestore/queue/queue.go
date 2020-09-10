package queue

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/balesz/go/firebase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// New creates a new queue
func New(statePath string, forceRunPath string) (queue Queue, err error) {
	var pathRegexp = regexp.MustCompile(`^\w+(?:/[\w\$\-]+)*$`)

	if statePath == "" {
		err = fmt.Errorf("The statePath parameter is empty")
		return
	} else if !pathRegexp.MatchString(statePath) {
		err = fmt.Errorf("The statePath parameter is invalid")
		return
	} else if len(strings.Split(strings.TrimSpace(strings.Trim(statePath, "/")), "/"))%2 != 0 {
		err = fmt.Errorf("The statePath parameter is not a document path")
		return
	}

	if forceRunPath == "" {
		err = fmt.Errorf("The forceRunPath parameter is empty")
		return
	} else if !pathRegexp.MatchString(forceRunPath) {
		err = fmt.Errorf("The forceRunPath parameter is invalid")
		return
	} else if len(strings.Split(strings.TrimSpace(strings.Trim(forceRunPath, "/")), "/"))%2 != 0 {
		err = fmt.Errorf("The forceRunPath parameter is not a document path")
		return
	}

	queue = Queue{forceRunPath: forceRunPath, statePath: statePath}

	return
}

// Processor creates a new processor instance
func (queue Queue) Processor(workers ...Worker) (proc Processor, err error) {
	if queue.statePath == "" {
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
func (processor Processor) Process(ctx context.Context, id string) (err error) {
	firebase.InitializeClients()

	var process = processor.createProcess(ctx, id, 0)

	if err = process.start(); err != nil {
		err = fmt.Errorf("process.start: %v", err)
		return
	}

	if err = process.handle(); err != nil {
		err = fmt.Errorf("process.handle: %v", err)
	}

	if er := process.stop(); er != nil && err == nil {
		err = fmt.Errorf("process.stop: %v", er)
		return
	} else if er != nil && err != nil {
		log.Println(fmt.Errorf("process.stop: %v", er))
		return
	} else if er == nil && err != nil {
		return
	}

	if err = process.forceRun(); err != nil {
		err = fmt.Errorf("process.forceRun: %v", err)
		return
	}

	return
}

func (processor Processor) createProcess(ctx context.Context, id string, idx int) process {
	return process{ctx: ctx, processID: id, processor: processor, worker: processor.workers[idx]}
}

func (process process) start() error {
	var (
		ctx         = process.ctx
		doc         = firebase.Firestore.Doc(process.processor.queue.statePath)
		forceRunRef = firebase.Firestore.Doc(process.processor.queue.forceRunPath)
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
				ForceRunRef: forceRunRef,
				IsRunning:   true,
				LastRunID:   processID,
			})
		}

		var state State
		err = snap.DataTo(&state)
		if err != nil {
			return fmt.Errorf("snap.DataTo: %v", err)
		}

		if state.Disabled {
			return fmt.Errorf("The queue is disabled")
		}

		if state.IsRunning {
			return fmt.Errorf("The queue is running")
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
		} else if state.ForceRunRef == nil {
			return fmt.Errorf("Missing forceRunRef field")
		}

		return tran.Set(state.ForceRunRef, ForceRunState{
			QueueStateRef: doc,
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}
