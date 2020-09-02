package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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

// DataTo convert FSEventValue to the given type
func (val FSEventValue) DataTo(dest interface{}) {
	var data = map[string]interface{}{}
	for k, v := range val.Fields {
		data[k] = v.Value()
	}
	encoded, _ := json.Marshal(data)
	json.Unmarshal(encoded, dest)
}

// FSEventField FSEventField
type FSEventField struct {
	AsString    *string    `json:"stringValue,omitempty"`
	AsTimestamp *time.Time `json:"timestampValue,omitempty"`
}

func (field FSEventField) String() string {
	return fmt.Sprintf("%v", field.Value())
}

// Value gets the dynamic value of the field
func (field FSEventField) Value() interface{} {
	if field.AsString != nil {
		return *field.AsString
	} else if field.AsTimestamp != nil {
		return *field.AsTimestamp
	}
	return nil
}

// FSEventUpdateMask FSEventUpdateMask
type FSEventUpdateMask struct {
	FieldPaths []string `json:"fieldPaths"`
}

func (mask FSEventUpdateMask) String() string {
	format := "EventUpdateMask(FieldPaths: %v)"
	return fmt.Sprintf(format, mask.FieldPaths)
}

// MockEvent MockEvent
type MockEvent struct {
	_new     map[string]interface{}
	_old     map[string]interface{}
	New      interface{}
	Old      interface{}
	Resource string
	Trigger  string
}

// CreateContext creates context for testing
func (evnt *MockEvent) CreateContext(base context.Context) (ctx context.Context, event FSEvent) {
	evnt._new = evnt.unmarshal(evnt.New)
	evnt._old = evnt.unmarshal(evnt.Old)
	ctx = metadata.NewContext(base, &metadata.Metadata{
		EventID:   time.Now().UTC().Format(time.RFC3339),
		EventType: evnt.Trigger,
		Resource:  &metadata.Resource{RawPath: evnt.Resource},
		Timestamp: time.Now(),
	})
	event = FSEvent{
		OldValue:   evnt.eventValue(evnt._old),
		Value:      evnt.eventValue(evnt._new),
		UpdateMask: evnt.eventUpdateMask(),
	}
	return
}

func (evnt MockEvent) unmarshal(data interface{}) map[string]interface{} {
	marshal, _ := json.Marshal(data)
	var unmarshal map[string]interface{}
	json.Unmarshal(marshal, &unmarshal)
	return unmarshal
}

func (evnt MockEvent) eventValue(doc map[string]interface{}) FSEventValue {
	var fields = map[string]FSEventField{}

	for key, value := range doc {
		if val, ok := value.(string); ok {
			fields[key] = FSEventField{AsString: &val}
		} else if val, ok := value.(time.Time); ok {
			fields[key] = FSEventField{AsTimestamp: &val}
		}
	}

	return FSEventValue{
		CreateTime: time.Now().Add(time.Duration(rand.Intn(72)) * time.Hour),
		Fields:     fields,
		Name:       evnt.Resource,
		UpdateTime: time.Now(),
	}
}

func (evnt MockEvent) eventUpdateMask() FSEventUpdateMask {
	var keys = map[string]bool{}

	for key, val := range evnt._old {
		if evnt._new[key] != val {
			keys[key] = true
		}
	}

	for key, val := range evnt._new {
		if evnt._old[key] != val {
			keys[key] = true
		}
	}

	fieldPaths := []string{}
	for key := range keys {
		fieldPaths = append(fieldPaths, key)
	}

	return FSEventUpdateMask{FieldPaths: fieldPaths}
}
