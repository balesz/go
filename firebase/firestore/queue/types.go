package queue

import (
	"context"
	"time"
)

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	CreateTime time.Time   `json:"createTime"`
	Fields     interface{} `json:"fields"`
	Name       string      `json:"name"`
	UpdateTime time.Time   `json:"updateTime"`
}

// State is the type of the queue state holder
type State struct {
	ForceRun  bool      `json:"forceRun"`
	IsRunning bool      `json:"isRunning"`
	LastRun   time.Time `json:"lastRun"`
}

// Change is represent a change on a queue state document
type Change struct {
	OldValue State
	NewValue State
}

// Handler type define the interface of queue handling functions
type Handler interface {
	Handle(context.Context)
	NeedForceRun(context.Context)
}
