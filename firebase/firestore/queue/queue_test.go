package queue

import (
	"context"
	"testing"

	"cloud.google.com/go/firestore"

	"github.com/balesz/go/env"
	"github.com/balesz/go/firebase"
)

var (
	executionID = "helloWorld"
	//executionID = time.Now().UTC().Format(time.RFC3339)
)

func TestEnvironment(t *testing.T) {
	var ctx = context.Background()
	if err := env.Init("game", "../../../.env"); err != nil {
		t.Error(err)
	} else if err := firebase.InitializeClients(); err != nil {
		t.Error(err)
	} else if _, err := firebase.Firestore.Doc("test/test").Get(ctx); err != nil {
		t.Error(err)
	}
}

func TestNew(t *testing.T) {
	var want, want1, want2 string
	want = "The statePath parameter is empty"
	if _, got := New("", ""); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
	want = "The statePath parameter is invalid"
	if _, got := New("/test/test/", ""); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
	want = "The forceRunPath parameter is invalid"
	if _, got := New("/test/test", "/hello/"); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
	want1, want2 = "test/$queue_state", "test/$queue_state/force/$queue_force_run"
	if got, _ := New("/test", ""); want1 != got.statePath {
		t.Errorf("%v != %v", want1, got.statePath)
	} else if want2 != got.forceRunPath {
		t.Errorf("%v != %v", want2, got.forceRunPath)
	}
	want1, want2 = "test/lobby", "test/lobby/force/$queue_force_run"
	if got, _ := New("/test/lobby", ""); want1 != got.statePath {
		t.Errorf("%v != %v", want1, got.statePath)
	} else if want2 != got.forceRunPath {
		t.Errorf("%v != %v", want2, got.forceRunPath)
	}
}

func TestStart(t *testing.T) {
	var ctx = context.Background()
	env.Init("game", "../../../.env")
	firebase.InitializeClients()

	queue, _ := New("test", "test")
	processor, _ := queue.Processor(mockHandler{})
	process := processor.createProcess(ctx, executionID, 0)

	if err := process.start(); err != nil {
		t.Error(err)
	}
}

func TestHandle(t *testing.T) {
	var ctx = context.Background()
	env.Init("game", "../../../.env")
	firebase.InitializeClients()

	queue, _ := New("test", "test")
	processor, _ := queue.Processor(mockHandler{})
	process := processor.createProcess(ctx, executionID, 0)

	if err := process.handle(); err != nil {
		t.Error(err)
	}
}

func TestStop(t *testing.T) {
	var ctx = context.Background()
	env.Init("game", "../../../.env")
	firebase.InitializeClients()

	queue, _ := New("test", "test")
	processor, _ := queue.Processor(mockHandler{})
	process := processor.createProcess(ctx, executionID, 0)

	if err := process.stop(); err != nil {
		t.Error(err)
	}
}

func TestForceRun(t *testing.T) {
	var ctx = context.Background()
	env.Init("game", "../../../.env")
	firebase.InitializeClients()

	queue, _ := New("test", "test")
	processor, _ := queue.Processor(mockHandler{})
	process := processor.createProcess(ctx, executionID, 0)

	if err := process.forceRun(); err != nil {
		t.Error(err)
	}
}

type mockHandler struct{}

func (handler mockHandler) Execute(ctx context.Context, tran *firestore.Transaction) error {
	doc := firebase.Firestore.Doc("test/test")
	err := tran.Set(doc, map[string]interface{}{"test": firestore.ServerTimestamp})
	if err != nil {
		return err
	}
	return nil
}

func (handler mockHandler) NeedForceExec(ctx context.Context, tran *firestore.Transaction) bool {
	return true
}
