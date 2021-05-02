package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go/v4"
	"github.com/mthorning/go-sso/config"
	"google.golang.org/api/option"
	"log"
)

type Config struct {
	GoogleApplicationCredentials string `required:"true" split_words:"true"`
}

type Update = firestore.Update

var Users *firestore.CollectionRef

func init() {
	var conf Config
	config.SetConfig(&conf)

	ctx := context.Background()
	sa := option.WithCredentialsFile(conf.GoogleApplicationCredentials)

	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	Users = client.Collection("users")
}
