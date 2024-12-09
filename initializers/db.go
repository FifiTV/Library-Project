package initializers

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var Client *firestore.Client

func ConnectToDb(ctx context.Context) error {

	opt := option.WithCredentialsFile("serviceAccountKey.json")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing app: %v", err)
	}

	Client, err = app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("error initializing Firestore: %v", err)
	}
	log.Println("Connected to Firestore")
	return nil
}
