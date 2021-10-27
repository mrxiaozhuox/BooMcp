package library

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataBase struct {
	client *mongo.Client
}

type UserInfo struct {
	username string
	email    string
	sex      int
	about    string
	password string
	salt     string
	status   int
}

func MongoConnect(uri string) (*DataBase, error) {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return &DataBase{}, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return &DataBase{}, err
	}

	return &DataBase{
		client: client,
	}, nil
}

func (db DataBase) Ping() bool {
	err := db.client.Ping(context.TODO(), nil)
	return err == nil
}

func (mongo DataBase) Register() error {
	_ = mongo.client.Database("fkycmp")
	return nil
}
