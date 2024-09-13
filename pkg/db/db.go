package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// mongo.Connect return mongo.Client method
	uri := "mongodb://localhost:27017/"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Error: " + err.Error())
	}

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error: " + err.Error())
	}
	log.Println("Connected to MongoDB")
	return client
}

// Client instance
var DB  *mongo.Client = ConnectDB()
var Name = "bucket"
var Opt = options.GridFSBucket().SetName(Name)


// Getting database collections
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("ImgAPI").Collection(collectionName)
	return collection
}
