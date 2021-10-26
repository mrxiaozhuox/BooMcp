package library

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataBase struct {
	client *mongo.Client
}

func MongoConnect(uri string) (DataBase, error) {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return DataBase{}, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return DataBase{}, err
	}

	return DataBase{
		client: client,
	}, nil
}
