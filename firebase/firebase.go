package firebase

import (
	"context"
	"fmt"
	"os"
	"strings"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	auth "firebase.google.com/go/auth"
	database "firebase.google.com/go/db"
)

// App is the default Firebase app instance
var App *firebase.App

// Auth is the default Authentication instance
var Auth *auth.Client

// Database is the default Realtime Database instance
var Database *database.Client

// Firestore is the default Firestore instance
var Firestore *firestore.Client

// InitializeClients initialize Firebase app and clients
func InitializeClients() error {
	if App != nil {
		return nil
	}
	var err error
	ctx := context.Background()
	App, err = firebase.NewApp(ctx, nil)
	if err != nil {
		return fmt.Errorf("firebase.NewApp: %v", err)
	}
	Auth, err = App.Auth(ctx)
	if err != nil {
		return fmt.Errorf("App.Auth: %v", err)
	}
	Firestore, err = App.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("App.Firestore: %v", err)
	}
	Database, err = App.Database(ctx)
	if err != nil {
		return fmt.Errorf("App.Database: %v", err)
	}
	return nil
}

// CheckEnvironment checks the environment
func CheckEnvironment() error {
	if os.Getenv("FIREBASE_CONFIG") == "" {
		return fmt.Errorf("FIREBASE_CONFIG environment not found")
	}
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS environment not found")
	}
	return nil
}

// DBPath creates a valid Realtime Database path from the given parts
func DBPath(parts ...string) string {
	return "/" + strings.Join(parts, "/")
}

// FSPath creates a valid Realtime Database path from the given parts
func FSPath(parts ...string) string {
	return strings.Join(parts, "/")
}
