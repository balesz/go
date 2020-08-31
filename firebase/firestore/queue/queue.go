package queue

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
)

// DocumentPath returns the queue handler document of the given collection
func DocumentPath(collection string) string {
	return collection + "/$queue"
}

// Start starts the queue handling process
func Start(ctx context.Context, collection string) error {
	projectID := os.Getenv("PROJECT_ID")
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	doc := client.Doc(DocumentPath(collection))

	return client.RunTransaction(ctx, func(ctx context.Context, tran *firestore.Transaction) error {
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
	})
}

// Handle handle the queue on the given collection
func (event *FirestoreEvent) Handle(ctx context.Context, collection string, handler Handler) error {
	if !event.needHandle() {
		return fmt.Errorf("The queue state is invalid")
	}

	projectID := os.Getenv("PROJECT_ID")
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	doc := client.Doc(DocumentPath(collection))

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

func (event *FirestoreEvent) needHandle() bool {
	change := event.getChange()
	if change.NewValue.ForceRun {
		return true
	}
	return false
}

func (event *FirestoreEvent) getChange() Change {
	return Change{}
}
