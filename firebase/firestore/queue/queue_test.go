package queue

import (
	"context"
	"testing"

	"cloud.google.com/go/firestore"
)

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

type mockHandler struct{}

func (handler mockHandler) Handle(ctx context.Context, tran *firestore.Transaction) error {
	return nil
}

func (handler mockHandler) NeedForceRun(ctx context.Context, tran *firestore.Transaction) bool {
	return false
}
