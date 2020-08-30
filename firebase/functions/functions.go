package functions

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"cloud.google.com/go/functions/metadata"
)

// LogContext is logging the context of Cloud Function
func LogContext(ctx context.Context, event interface{}) {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return
	}

	log.Printf("Environment: %v", os.Environ())
	log.Printf("Event: %v", event)
	log.Printf("Metadata: %v", meta)
	log.Printf("Resource: %v", meta.Resource)

	res := meta.Resource
	log.Printf("Metadata{EventId: %v, EventType: %v, Timestamp: %v}", meta.EventID, meta.EventType, meta.Timestamp)
	log.Printf("Resource{Service: %v, Name: %v, Type: %v, RawPath: %v}", res.Service, res.Name, res.Type, res.RawPath)
}

// GetFirestorePathFromResource is getting the document path from Metadata.Resource
func GetFirestorePathFromResource(ctx context.Context) (string, error) {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return "", err
	}
	rx := regexp.MustCompile(`.*\(default\)/documents/(.*)`)
	if !rx.MatchString(meta.Resource.RawPath) {
		return "", fmt.Errorf("getFirestorePath: Invalid RawPath")
	}
	path := rx.ReplaceAllString(meta.Resource.RawPath, "$1")
	log.Printf("getFirestorePath(%v)", path)
	return path, nil
}

// GetDatabasePathFromResource is getting the database ref path from Metadata.Resource
func GetDatabasePathFromResource(ctx context.Context) (string, error) {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return "", err
	}
	rx := regexp.MustCompile(`projects/_/instances/.*/refs/(.*)`)
	if !rx.MatchString(meta.Resource.RawPath) {
		return "", fmt.Errorf("GetDatabasePathFromResource: Invalid RawPath")
	}
	path := rx.ReplaceAllString(meta.Resource.RawPath, "/$1")
	log.Printf("GetDatabasePathFromResource(%v)", path)
	return path, nil
}
