package firebase

import (
	"context"
	"log"
	"regexp"
	"testing"
)

func init() {
	err := CheckEnvironment()
	if err != nil {
		log.Fatalf("firebase.CheckEnvironment: %v", err)
	}
	err = InitializeClients()
	if err != nil {
		log.Fatalf("firebase.InitializeClients: %v", err)
	}
}

const path = "test/test"

type Test struct {
	Foo   string `firestore:"foo"`
	Hello string `firestore:"hello"`
}

func TestFirestore(test *testing.T) {
	ctx := context.Background()
	doc := Firestore.Doc(path)
	snap, err := doc.Get(ctx)
	if err != nil {
		test.Error(err)
	}

	var dataObj Test
	snap.DataTo(&dataObj)

	log.Println(dataObj)
}

func TestRealtimeDatabase(test *testing.T) {
	ctx := context.Background()
	var play Test
	Database.NewRef(path).Get(ctx, &play)
	log.Println(play)
}

func TestGetFirestorePath(test *testing.T) {
	var input = "projects/gamerpro-game/databases/(default)/documents/test/test"

	rx := regexp.MustCompile(`.*\(default\)\/documents\/(.*)`)
	path := rx.ReplaceAllString(input, "$1")

	if path != "test/test" {
		test.Errorf("the regexp is not valid")
	} else {
		log.Printf("TestGetFirestorePath: %v", path)
	}
}

func TestGetDatabasePath(test *testing.T) {
	var input = "projects/_/instances/gamerpro-game/refs/test/test"

	rx := regexp.MustCompile(`projects/_/instances/.*/refs/(.*)`)
	path := rx.ReplaceAllString(input, "$1")

	if path != "test/test" {
		test.Errorf("the regexp is not valid")
	} else {
		log.Printf("TestGetDatabasePath: %v", path)
	}
}
