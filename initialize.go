package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"

	"google.golang.org/api/option"
)

func initializeApp() (*firebase.App, error) {
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	return app, nil
}

func registerUser(email, password, username string) error {
	ctx := context.Background()

	// Initialize Firebase App
	app, err := initializeApp()
	if err != nil {
		return err
	}

	// Initialize Firebase Auth Client
	authClient, err := app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("error getting Auth client: %v", err)
	}

	// Create a new Firebase Authentication user
	params := (&auth.UserToCreate{}).
		Email(email).
		Password(password).
		DisplayName(username)
	userRecord, err := authClient.CreateUser(ctx, params)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	fmt.Printf("Successfully created user: %v\n", userRecord.UID)

	// Initialize Firestore Client
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("error getting Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Save additional user information in Firestore
	_, err = firestoreClient.Collection("users").Doc(userRecord.UID).Set(ctx, map[string]interface{}{
		"username": username,
		"email":    email,
		"created":  firestore.ServerTimestamp,
	})
	if err != nil {
		return fmt.Errorf("error saving user to Firestore: %v", err)
	}

	fmt.Printf("Successfully registered user %s with Firestore.\n", username)
	return nil
}
