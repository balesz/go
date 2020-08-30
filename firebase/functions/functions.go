package functions

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"cloud.google.com/go/functions/metadata"
)

var rxFirestorePath = regexp.MustCompile(`.*\(default\)/documents/(.*)`)

var rxDatabasePath = regexp.MustCompile(`projects/_/instances/.*/refs/(.*)`)

// InitializeLog initialize log for Cloud Functions
func InitializeLog() {
	log.SetFlags(log.Flags() &^ log.Ltime &^ log.Ldate)
}

// LogContext is logging the context of Cloud Function
func LogContext(ctx context.Context, event interface{}) {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return
	}
	log.Printf("Metadata: %v", meta)
	log.Printf("Resource: %v", meta.Resource)
	log.Printf("Event: %v", event)
}

// GetPath gets the path of the resource
func GetPath(ctx context.Context) (string, error) {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return "", err
	}
	var path string
	if rxFirestorePath.MatchString(meta.Resource.RawPath) {
		path = rxFirestorePath.ReplaceAllString(meta.Resource.RawPath, "$1")
	} else if rxDatabasePath.MatchString(meta.Resource.RawPath) {
		path = rxDatabasePath.ReplaceAllString(meta.Resource.RawPath, "/$1")
	} else {
		return "", fmt.Errorf("Invalid Metadata.Resource.RawPath")
	}
	return path, nil
}
