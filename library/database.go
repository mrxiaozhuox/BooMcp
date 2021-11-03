package library

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gobuffalo/packr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/gomail.v2"
)

type DataBase struct {
	client *mongo.Client
	config GeneralConfig
	packer packr.Box
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

func MongoConnect(config GeneralConfig, pack packr.Box) (*DataBase, error) {

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
		packer: pack,
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
	err := collection.FindOne(
		context.TODO(),
		bson.D{
			{
				Key:   "email",
				Value: user.Email,
			},
		},
	).Decode(&temp)

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

			mail := gomail.NewMessage()

			mail.SetHeader("From", mongo.config.EmailConfig.Username)
			mail.SetHeader("To", user.Email)
			mail.SetHeader("Subject", "账号邮箱验证「 "+mongo.config.SiteName+" 」")

			// 加载相应的数据模板文件
			templ, err := mongo.packer.FindString("email/check-email.tmpl")
			if err != nil {
				fmt.Println("Email发送模板不存在。")
			}

			mail.SetBody("text/html", templ)

			status, err := SendEmail(mongo.config.EmailConfig, mail)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(0)
			}
			log.Println(status)
		}

		return true, nil
	}

	return false, nil
}
