package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	DB     *mongo.Database

	UsersCollection *mongo.Collection
	PostsCollection *mongo.Collection
)

func Init(ctx context.Context, uri, dbName string) error {
	cl, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err = cl.Connect(ctx2)
	if err != nil {
		return err
	}
	// ping
	ctx3, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()
	if err := cl.Ping(ctx3, nil); err != nil {
		return err
	}

	Client = cl
	DB = cl.Database(dbName)
	UsersCollection = DB.Collection("users")
	PostsCollection = DB.Collection("posts")
	return nil
}
