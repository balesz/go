package functions

import (
	"fmt"
	"time"
)

// FSTriggerCreate : Triggered when a document is written to for the first time.
const FSTriggerCreate = "providers/cloud.firestore/eventTypes/document.create"

// FSTriggerDelete :  Triggered when a document with data is deleted.
const FSTriggerDelete = "providers/cloud.firestore/eventTypes/document.delete"

// FSTriggerUpdate : Triggered when a document already exists and has any value changed.
const FSTriggerUpdate = "providers/cloud.firestore/eventTypes/document.update"

// FSTriggerWrite : Triggered when a document is created, updated or deleted.
const FSTriggerWrite = "providers/cloud.firestore/eventTypes/document.write"

// FSString : String typed value
type FSString struct {
	Value string `json:"stringValue"`
}

// FSTimestamp : Timestamp typed value
type FSTimestamp struct {
	Value time.Time `json:"timestampValue"`
}

func (val FSString) String() string {
	return fmt.Sprintf("%v", val.Value)
}

func (val FSTimestamp) String() string {
	return fmt.Sprintf("%v", val.Value.UTC().Format(time.RFC3339))
}
