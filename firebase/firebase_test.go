package firebase

import (
	"context"
	"log"
	"regexp"
	"testing"

	"github.com/balesz/go/env"
)

func init() {
	env.Init("game", "../.env")
	err := CheckEnvironment()
	if err != nil {
		log.Fatalf("firebase.CheckEnvironment: %v", err)
	}
	err = InitializeClients()
	if err != nil {
		log.Fatalf("firebase.InitializeClients: %v", err)
	}
}

func TestMisc(test *testing.T) {}

func xTestFirestore(test *testing.T) {
	const path = "test/test"
	ctx := context.Background()

	doc := Firestore.Doc(path)
	snap, err := doc.Get(ctx)
	if err != nil {
		test.Error(err)
	}

	var dataObj struct {
		Foo   string `firestore:"foo"`
		Hello string `firestore:"hello"`
	}

	snap.DataTo(&dataObj)

	log.Println(dataObj)
}

func xTestRealtimeDatabase(test *testing.T) {
	const path = "test/test"
	ctx := context.Background()

	var play struct {
		Foo   string `firestore:"foo"`
		Hello string `firestore:"hello"`
	}

	Database.NewRef(path).Get(ctx, &play)
	log.Println(play)
}

func xTestGetFirestorePath(test *testing.T) {
	var input = "projects/gamerpro-game/databases/(default)/documents/test/test"

	rx := regexp.MustCompile(`.*\(default\)\/documents\/(.*)`)
	path := rx.ReplaceAllString(input, "$1")

	if path != "test/test" {
		test.Errorf("the regexp is not valid")
	} else {
		log.Printf("TestGetFirestorePath: %v", path)
	}
}

func xTestGetDatabasePath(test *testing.T) {
	var input = "projects/_/instances/gamerpro-game/refs/test/test"

	rx := regexp.MustCompile(`projects/_/instances/.*/refs/(.*)`)
	path := rx.ReplaceAllString(input, "$1")

	if path != "test/test" {
		test.Errorf("the regexp is not valid")
	} else {
		log.Printf("TestGetDatabasePath: %v", path)
	}
}
