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

	queue = Queue{ForceRunPath: forceRunPath, StatePath: statePath}

	return
}

// NewTask creates a new Task instance
func (queue Queue) NewTask(id string, worker Worker) (task Task, err error) {
	if queue.StatePath == "" {
		err = fmt.Errorf("Queue is not initialized")
	}
	task = Task{ID: id, queue: queue, worker: worker}
	return
}

// Dispatch method is execute the workers of the task
func (task Task) Dispatch(ctx context.Context) (err error) {
	if err = task.start(ctx); err != nil {
		err = fmt.Errorf("task.start: %v", err)
		return
	}

	if err = task.handle(ctx); err != nil {
		err = fmt.Errorf("task.handle: %v", err)
	}

	if er := task.stop(ctx); er != nil && err == nil {
		err = fmt.Errorf("task.stop: %v", er)
		return
	} else if er != nil && err != nil {
		log.Println(fmt.Errorf("task.stop: %v", er))
		return
	} else if er == nil && err != nil {
		return
	}

	if err = task.forceRun(ctx); err != nil {
		err = fmt.Errorf("task.forceRun: %v", err)
		return
	}

	return
}

func (task Task) start(ctx context.Context) error {
	var (
		stateRef    = firebase.Firestore.Doc(task.queue.StatePath)
		forceRunRef = firebase.Firestore.Doc(task.queue.ForceRunPath)
		maxAttempts = firestore.MaxAttempts(1)
	)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(stateRef)
		if err != nil && status.Code(err) != codes.NotFound {
			return fmt.Errorf("tran.Get: %v", err)
		}

		tran.Delete(forceRunRef)

		if !snap.Exists() {
			return tran.Create(stateRef, State{
				ForceRunRef: forceRunRef,
				IsRunning:   true,
				LastTaskID:  task.ID,
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

		return tran.Update(stateRef, []firestore.Update{
			{Path: "isRunning", Value: true},
			{Path: "lastRun", Value: firestore.ServerTimestamp},
			{Path: "lastTaskID", Value: task.ID},
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}

func (task Task) handle(ctx context.Context) error {
	var (
		stateRef    = firebase.Firestore.Doc(task.queue.StatePath)
		maxAttempts = firestore.MaxAttempts(5)
	)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(stateRef)
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
			return fmt.Errorf("The queue not running")
		} else if state.LastTaskID != task.ID {
			return fmt.Errorf("The current task is not this one")
		}

		return task.worker.Execute(ctx, tran)
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}

func (task Task) stop(ctx context.Context) error {
	var (
		stateRef    = firebase.Firestore.Doc(task.queue.StatePath)
		maxAttempts = firestore.MaxAttempts(5)
	)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(stateRef)
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
			return fmt.Errorf("The queue not running")
		} else if state.LastTaskID != task.ID {
			return fmt.Errorf("The current task is not this one")
		}

		return tran.Update(stateRef, []firestore.Update{
			{Path: "isRunning", Value: false},
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}

func (task Task) forceRun(ctx context.Context) error {
	var (
		stateRef    = firebase.Firestore.Doc(task.queue.StatePath)
		maxAttempts = firestore.MaxAttempts(2)
	)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		if err := task.worker.NeedForceExec(ctx, tran); err != nil {
			log.Printf("worker.NeedForceExec: %v", err)
			return nil
		}

		snap, err := tran.Get(stateRef)
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
			return fmt.Errorf("The queue is running")
		} else if state.LastTaskID != task.ID {
			return fmt.Errorf("The current task is not this one")
		} else if state.ForceRunRef == nil {
			return fmt.Errorf("Missing forceRunRef field")
		}

		return tran.Set(state.ForceRunRef, ForceRunState{
			QueueStateRef: stateRef,
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}
