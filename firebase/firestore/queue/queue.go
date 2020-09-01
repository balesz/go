package queue

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/balesz/go/firebase"
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

// Execute is execute the queue runner
func (runner Runner) Execute(ctx context.Context) error {
	firebase.InitializeClients()

	if err := runner.init(); err != nil {
		return fmt.Errorf("runner.init: %v", err)
	} else if err := runner.start(ctx); err != nil {
		return fmt.Errorf("runner.start: %v", err)
	} else if err := runner.handle(ctx); err != nil {
		return fmt.Errorf("runner.handle: %v", err)
	} else if err := runner.stop(ctx); err != nil {
		return fmt.Errorf("runner.stop: %v", err)
	} else if err := runner.needForceRun(ctx); err != nil {
		return fmt.Errorf("runner.needForceRun: %v", err)
	}

	return nil
}

func (runner *Runner) init() error {
	if runner._initialized {
		return fmt.Errorf("The queue is already initialized")
	} else if runner.ExecutionID == "" {
		return fmt.Errorf("The ExecutionID field is empty")
	} else if runner.Handler == nil {
		return fmt.Errorf("The Handler field is empty")
	} else if runner.Path == "" {
		return fmt.Errorf("The Path field is empty")
	} else if !regexp.MustCompile(`^/?\w+(?:/\w+)*$`).MatchString(runner.Path) {
		return fmt.Errorf("The Path field is invalid")
	}

	var path = strings.TrimPrefix(strings.TrimSpace(runner.Path), "/")
	parts := strings.Split(path, "/")
	if len(parts)%2 == 0 {
		parts = parts[:len(parts)-1]
	}

	runner._documentPath = strings.Join(append(parts, stateDocID), "/")

	runner._initialized = true

	return nil
}

func (runner Runner) start(ctx context.Context) error {
	maxAttempts := firestore.MaxAttempts(1)
	var doc = firebase.Firestore.Doc(runner._documentPath)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(doc)
		if err != nil {
			return fmt.Errorf("tran.Get: %v", err)
		}

		if !snap.Exists() {
			return tran.Create(doc, State{
				IsRunning: true,
				LastRunID: runner.ExecutionID,
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
			{Path: "lastRunID", Value: runner.ExecutionID},
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}

func (runner Runner) handle(ctx context.Context) error {
	maxAttempts := firestore.MaxAttempts(5)
	var doc = firebase.Firestore.Doc(runner._documentPath)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(doc)
		if err != nil {
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
		} else if state.LastRunID != runner.ExecutionID {
			return fmt.Errorf("The currently running is not this runner")
		}

		return runner.Handler.Handle(ctx, tran)
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}

func (runner Runner) stop(ctx context.Context) error {
	maxAttempts := firestore.MaxAttempts(5)
	var doc = firebase.Firestore.Doc(runner._documentPath)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(doc)
		if err != nil {
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
		} else if state.LastRunID != runner.ExecutionID {
			return fmt.Errorf("The currently running is not this runner")
		}

		return tran.Update(doc, []firestore.Update{
			{Path: "isRunning", Value: false},
			{Path: "lastRun", Value: time.Now()},
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}

func (runner Runner) needForceRun(ctx context.Context) error {
	maxAttempts := firestore.MaxAttempts(2)
	var doc = firebase.Firestore.Doc(runner._documentPath)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		if need := runner.Handler.NeedForceRun(ctx, tran); !need {
			return nil
		}

		snap, err := tran.Get(doc)
		if err != nil {
			return fmt.Errorf("tran.Get: %v", err)
		} else if !snap.Exists() {
			return fmt.Errorf("The queue state document not exists")
		}

		var state State
		err = snap.DataTo(&state)
		if err != nil {
			return fmt.Errorf("snap.DataTo: %v", err)
		} else if state.IsRunning {
			return fmt.Errorf("The queue runner running")
		}

		return tran.Update(doc, []firestore.Update{
			{Path: "forceRun", Value: time.Now()},
			{Path: "lastRun", Value: time.Now()},
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction, maxAttempts)
}
