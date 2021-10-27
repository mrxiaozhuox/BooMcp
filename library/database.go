package library

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
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
	collection := db.Collection("Users")

	// 检查是否已经被注册
	var temp UserInfo
	err := collection.FindOne(context.TODO(), bson.D{{"email", user.Email}}).Decode(&temp)

	if err != nil {

		// 数据没找到，可以插入
		_, err := collection.InsertOne(context.TODO(), user)
		if err != nil {
			return false, err
		}

		// 插入账号待验证信息
		if mongo.config.EmailConfig.Server != "" {
			// 不为空则说明配置了邮箱系统信息
			// 自动检测是否支持
		}

		return true, nil
	}

	return false, nil
}
