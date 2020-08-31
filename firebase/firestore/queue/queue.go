package queue

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/balesz/go/firebase"
	"github.com/balesz/go/firebase/functions"
)

const stateDocID = "/$queue"

// HandleQueueStateDocument handles queue state doc change
func HandleQueueStateDocument(ctx context.Context) bool {
	path, err := functions.GetPath(ctx)
	if err != nil || !strings.HasSuffix(path, stateDocID) {
		return false
	}
	return true
}

// Start starts the queue handling process
func Start(ctx context.Context) error {
	stateDocPath, _ := functions.GetPath(ctx)

	doc := firebase.Firestore.Doc(stateDocPath)

	transaction := func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(doc)
		if err != nil {
			return err
		}

		if !snap.Exists() {
			return tran.Create(doc, []firestore.Update{
				{Path: "forceRun", Value: true},
				{Path: "isRunning", Value: false},
				{Path: "lastRun", Value: nil},
			})
		}

		var state State
		err = snap.DataTo(&state)
		if err != nil {
			return err
		}

		if state.IsRunning || state.ForceRun {
			return nil
		}

		return tran.Update(doc, []firestore.Update{
			{Path: "forceRun", Value: true},
		})
	}

	return firebase.Firestore.RunTransaction(ctx, transaction)
}

// Handle handle the queue on the given collection
func Handle(ctx context.Context, collection string, handler Handler) error {
	projectID := os.Getenv("PROJECT_ID")
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	doc := client.Doc(documentPath(collection))

	return client.RunTransaction(ctx, func(ctx context.Context, tran *firestore.Transaction) error {
		snap, err := tran.Get(doc)
		if err != nil {
			return err
		}

		if !snap.Exists() {
			return fmt.Errorf("The queue state document not exists")
		}

		return nil
	})
}

func documentPath(collection string) string {
	return collection + stateDocID
}
