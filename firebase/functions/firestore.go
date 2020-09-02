package functions

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/functions/metadata"
)

// FSTriggerCreate : Triggered when a document is written to for the first time.
const FSTriggerCreate = "providers/cloud.firestore/eventTypes/document.create"

// FSTriggerDelete :  Triggered when a document with data is deleted.
const FSTriggerDelete = "providers/cloud.firestore/eventTypes/document.delete"

// FSTriggerUpdate : Triggered when a document already exists and has any value changed.
const FSTriggerUpdate = "providers/cloud.firestore/eventTypes/document.update"

// FSTriggerWrite : Triggered when a document is created, updated or deleted.
const FSTriggerWrite = "providers/cloud.firestore/eventTypes/document.write"

// FSEvent FSEvent
type FSEvent struct {
	OldValue   FSEventValue      `json:"oldValue"`
	Value      FSEventValue      `json:"value"`
	UpdateMask FSEventUpdateMask `json:"updateMask"`
}

func (event FSEvent) String() string {
	format := "Event(OldValue: %v, Value: %v, UpdateMask: %v)"
	return fmt.Sprintf(format, event.OldValue, event.Value, event.UpdateMask)
}

// FSEventValue FSEventValue
type FSEventValue struct {
	CreateTime time.Time               `json:"createTime"`
	Fields     map[string]FSEventField `json:"fields"`
	Name       string                  `json:"name"`
	UpdateTime time.Time               `json:"updateTime"`
}

func (val FSEventValue) String() string {
	format := "EventValue(CreateTime: %v, Fields: %v, Name: %v, UpdateTime: %v)"
	return fmt.Sprintf(format, val.CreateTime, val.Fields, val.Name, val.Name)
}

// FSEventField FSEventField
type FSEventField struct {
	AsString    *string    `json:"stringValue,omitempty"`
	AsTimestamp *time.Time `json:"timestampValue,omitempty"`
}

func (field FSEventField) String() string {
	if field.AsString != nil {
		return fmt.Sprintf("%v", *field.AsString)
	} else if field.AsTimestamp != nil {
		return fmt.Sprintf("%v", (*field.AsTimestamp).UTC().Format(time.RFC3339))
	}
	return fmt.Sprintf("<Unknown>")
}

// FSEventUpdateMask FSEventUpdateMask
type FSEventUpdateMask struct {
	FieldPaths []string `json:"fieldPaths"`
}

func (mask FSEventUpdateMask) String() string {
	format := "EventUpdateMask(FieldPaths: %v)"
	return fmt.Sprintf(format, mask.FieldPaths)
}

// MockChanges MockChanges
type MockChanges struct {
	Name string
	New  map[string]interface{}
	Old  map[string]interface{}
}

// FSMockContext FSMockContext
func FSMockContext(trigger string, changes MockChanges) (ctx context.Context, event FSEvent) {
	event = FSEvent{
		OldValue:   changes.eventValue(changes.Old),
		Value:      changes.eventValue(changes.New),
		UpdateMask: changes.eventUpdateMask(),
	}
	ctx = metadata.NewContext(context.Background(), &metadata.Metadata{
		EventID:   time.Now().UTC().Format(time.RFC3339),
		EventType: trigger,
		Resource:  &metadata.Resource{RawPath: changes.Name},
		Timestamp: time.Now(),
	})
	return
}

func (chg MockChanges) eventValue(doc map[string]interface{}) FSEventValue {
	var fields = map[string]FSEventField{}

	for key, value := range doc {
		if val, ok := value.(string); ok {
			fields[key] = FSEventField{AsString: &val}
		} else if val, ok := value.(time.Time); ok {
			fields[key] = FSEventField{AsTimestamp: &val}
		}
	}

	return FSEventValue{
		CreateTime: time.Now(),
		Fields:     fields,
		Name:       chg.Name,
		UpdateTime: time.Now(),
	}
}

func (chg MockChanges) eventUpdateMask() FSEventUpdateMask {
	var keys = map[string]bool{}

	for key, val := range chg.Old {
		if chg.New[key] != val {
			keys[key] = true
		}
	}

	for key, val := range chg.New {
		if chg.Old[key] != val {
			keys[key] = true
		}
	}

	fieldPaths := []string{}
	for key := range keys {
		fieldPaths = append(fieldPaths, key)
	}

	return FSEventUpdateMask{FieldPaths: fieldPaths}
}
