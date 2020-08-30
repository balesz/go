package firebase

import (
	"context"
	"fmt"
	"log"
	"os"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	database "firebase.google.com/go/db"
)

// App is the default Firebase app instance
var App *firebase.App

// Firestore is the default Firestore instance
var Firestore *firestore.Client

// Database is the default Realtime Database instance
var Database *database.Client

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

// InitializeLog initialize log for Google Cloud Logging
func InitializeLog() {
	log.SetFlags(log.Flags() &^ log.Ltime &^ log.Ldate)
}
