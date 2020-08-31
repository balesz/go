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

// FSValueString : String typed field
type FSValueString struct {
	StringValue string `json:"stringValue"`
}

// FSValueTimestamp : Timestamp typed field
type FSValueTimestamp struct {
	TimestampValue time.Time `json:"timestampValue"`
}

func (val *FSValueString) String() string {
	return fmt.Sprintf("%v", val.StringValue)
}

func (val *FSValueTimestamp) String() string {
	return fmt.Sprintf("%v", time.Now().UTC().Format(time.RFC3339))
}
