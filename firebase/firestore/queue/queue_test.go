package queue

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/firestore"

	"github.com/balesz/go/env"
	"github.com/balesz/go/firebase"
)

var (
	taskID = "helloWorld"
	//taskID = time.Now().UTC().Format(time.RFC3339)
)

func TestEnvironment(t *testing.T) {
	var ctx = context.Background()
	if _, err := env.Init("game", "../../../.env"); err != nil {
		t.Error(err)
	} else if err := firebase.InitializeClients(); err != nil {
		t.Error(err)
	} else if _, err := firebase.Firestore.Doc("test/test").Get(ctx); err != nil {
		t.Error(err)
	}
}

func TestNew(t *testing.T) {
	var want string
	want = "The statePath parameter is empty"
	if _, got := New("", ""); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
	want = "The statePath parameter is invalid"
	if _, got := New("/test/test", ""); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
	want = "The statePath parameter is not a document path"
	if _, got := New("test/test/test", ""); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
}

func TestStart(t *testing.T) {
	var ctx = context.Background()
	env.Init("game", "../../../.env")
	firebase.InitializeClients()

	queue, _ := New("test/--queue-state--", "test/--force-run--")
	task, _ := queue.NewTask(taskID, mockHandler{})

	if err := task.start(ctx); err != nil {
		t.Error(err)
	}
}

func TestHandle(t *testing.T) {
	var ctx = context.Background()
	env.Init("game", "../../../.env")
	firebase.InitializeClients()

	queue, _ := New("test/--queue-state--", "test/--force-run--")
	task, _ := queue.NewTask(taskID, mockHandler{})

	if err := task.handle(ctx); err != nil {
		t.Error(err)
	}
}

func TestStop(t *testing.T) {
	var ctx = context.Background()
	env.Init("game", "../../../.env")
	firebase.InitializeClients()

	queue, _ := New("test/--queue-state--", "test/--force-run--")
	task, _ := queue.NewTask(taskID, mockHandler{})

	if err := task.stop(ctx); err != nil {
		t.Error(err)
	}
}

func TestForceRun(t *testing.T) {
	var ctx = context.Background()
	env.Init("game", "../../../.env")
	firebase.InitializeClients()

	queue, _ := New("test/--queue-state--", "test/--force-run--")
	task, _ := queue.NewTask(taskID, mockHandler{})

	if err := task.forceRun(ctx); err != nil {
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

func (handler mockHandler) NeedForceExec(ctx context.Context, tran *firestore.Transaction) error {
	return fmt.Errorf("no need to force execute")
}
