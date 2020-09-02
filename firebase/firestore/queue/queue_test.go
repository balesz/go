package queue

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/balesz/go/firebase"
	"github.com/balesz/go/test"
)

var (
	//executionID = "helloWorld"
	executionID = time.Now().UTC().Format(time.RFC3339)
)

func TestEnvironment(t *testing.T) {
	var ctx = context.Background()
	if err := test.InitEnvironment("game", "../../../.env"); err != nil {
		t.Error(err)
	} else if err := firebase.InitializeClients(); err != nil {
		t.Error(err)
	} else if _, err := firebase.Firestore.Doc("test/test").Get(ctx); err != nil {
		t.Error(err)
	}
}

func TestIsStatePath(t *testing.T) {
	if want, got := false, IsStatePath("/lobby/helloUser"); want != got {
		t.Errorf("%v != %v", want, got)
	} else if want, got := true, IsStatePath("/lobby/$queue"); want != got {
		t.Errorf("%v != %v", want, got)
	}
}

func TestInit(t *testing.T) {
	var runner = Runner{}
	if want, got := "The ExecutionID field is empty", runner.init(); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
	runner.ExecutionID = "executuionID"
	if want, got := "The Handler field is empty", runner.init(); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
	runner.Handler = mockHandler{}
	if want, got := "The Path field is empty", runner.init(); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
	runner.Path = "/lobby/userID/"
	if want, got := "The Path field is invalid", runner.init(); want != got.Error() {
		t.Errorf("%v != %v", want, got)
	}
	runner.Path = "/lobby/userID"
	if got := runner.init(); nil != got {
		t.Errorf("%v != %v", nil, got)
	} else if !runner._initialized {
		t.Errorf("Runner is not initialized")
	} else if runner._documentPath != "lobby/$queue" {
		t.Errorf("Invalid document path")
	}
}

func TestStart(t *testing.T) {
	var ctx = context.Background()
	test.InitEnvironment("game", "../../../.env")
	firebase.InitializeClients()

	var execID = executionID
	var runner = Runner{ExecutionID: execID, Handler: mockHandler{}, Path: "test"}

	if err := runner.init(); err != nil {
		t.Error(err)
	} else if err := runner.start(ctx); err != nil {
		t.Error(err)
	}
}

func TestHandle(t *testing.T) {
	var ctx = context.Background()
	test.InitEnvironment("game", "../../../.env")
	firebase.InitializeClients()

	var execID = executionID
	var runner = Runner{ExecutionID: execID, Handler: mockHandler{}, Path: "test"}

	if err := runner.init(); err != nil {
		t.Error(err)
	} else if err := runner.handle(ctx); err != nil {
		t.Error(err)
	}
}

func TestStop(t *testing.T) {
	var ctx = context.Background()
	test.InitEnvironment("game", "../../../.env")
	firebase.InitializeClients()

	var execID = executionID
	var runner = Runner{ExecutionID: execID, Handler: mockHandler{}, Path: "test"}

	if err := runner.init(); err != nil {
		t.Error(err)
	} else if err := runner.stop(ctx); err != nil {
		t.Error(err)
	}
}

func TestForceRun(t *testing.T) {
	var ctx = context.Background()
	test.InitEnvironment("game", "../../../.env")
	firebase.InitializeClients()

	var execID = executionID
	var runner = Runner{ExecutionID: execID, Handler: mockHandler{}, Path: "test"}

	if err := runner.init(); err != nil {
		t.Error(err)
	} else if err := runner.forceRun(ctx); err != nil {
		t.Error(err)
	}
}

type mockHandler struct{}

func (handler mockHandler) Handle(ctx context.Context, tran *firestore.Transaction) error {
	doc := firebase.Firestore.Doc("test/test")
	err := tran.Set(doc, map[string]interface{}{"test": firestore.ServerTimestamp})
	if err != nil {
		return err
	}
	return nil
}

func (handler mockHandler) NeedForceRun(ctx context.Context, tran *firestore.Transaction) bool {
	return true
}
