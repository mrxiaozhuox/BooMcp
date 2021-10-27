package library

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataBase struct {
	client *mongo.Client
	config GeneralConfig
}

type UserInfo struct {
	Username string
	Email    string
	Sex      int
	About    string
	Password string
	Salt     string
	Status   int
	Level    int
}

func MongoConnect(config GeneralConfig) (*DataBase, error) {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.MongoDbURI))
	if err != nil {
		return &DataBase{}, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return &DataBase{}, err
	}

	return &DataBase{
		client: client,
		config: config,
	}, nil
}

func (db DataBase) Ping() bool {
	err := db.client.Ping(context.TODO(), nil)
	return err == nil
}

// 用户注册命令
func (mongo DataBase) Register(user UserInfo) (bool, error) {

	db := mongo.client.Database("fkycmp")

	_, err := db.Collection("Users").InsertOne(context.TODO(), user)

	if err != nil {
		return false, err
	}

	return true, nil
}
