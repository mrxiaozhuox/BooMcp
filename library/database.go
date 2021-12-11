package library

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gobuffalo/packr/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/gomail.v2"
)

const DATABASENAME string = "boomcp"

type DataBase struct {
	client *mongo.Client
	config GeneralConfig
	packer *packr.Box
}

type UserInfo struct {
	Id       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Username string
	Email    string
	About    string
	Password string
	Salt     string
	Status   int
	Level    int
	Regtime  primitive.DateTime
	Initacc  bool
	Mcsmpwd  map[string]string
	Mcsmid   string
}

type TokenStruct struct {
	Token     string
	Target    interface{}
	Operation string
}

func MongoConnect(config GeneralConfig, pack *packr.Box) (*DataBase, error) {

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

func (mongo DataBase) InitDataBase() error {

	if !mongo.Ping() {
		return errors.New("数据库连接错误")
	}

	// db := mongo.client.Database(DATABASENAME)
	// userCollect := db.Collection("Users")

	randBytes := make([]byte, 10/2)
	_, err := rand.Read(randBytes)
	if err != nil {
		return err
	}

	salt := fmt.Sprintf("%x", randBytes)

	// 默认密码为 mrxzx.info
	password := MetaPassword("mrxzx.info", salt)

	// 初始化系统第一个管理员用户
	// InitAcc 则代表这个账号允许在不经过邮箱验证的情况下更新一次个人信息（可以换邮箱）
	user := UserInfo{
		Username: "admin",
		Email:    "admin@mrxzx.info",
		Salt:     salt,
		Password: password,
		About:    "我就是最无敌的存在！",
		Status:   0,
		Level:    2,
		Initacc:  true,
		Regtime:  primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err = mongo.Register(user)
	return err
}

func (mongo DataBase) Ping() bool {
	err := mongo.client.Ping(context.TODO(), nil)
	return err == nil
}

func (mongo DataBase) SetTempData(dtype string, data interface{}, checker interface{}) (bool, error) {

	db := mongo.client.Database(DATABASENAME)
	collection := db.Collection("TempData")

	// 查找是否有重复的 checker
	einfo := collection.FindOne(context.TODO(), bson.D{
		{
			Key:   "checker",
			Value: checker,
		},
	}).Err()

	var err error

	value := bson.D{
		{
			Key:   "type",
			Value: dtype,
		},
		{
			Key:   "data",
			Value: data,
		},
		{
			Key:   "checker",
			Value: checker,
		},
	}

	if einfo == nil {
		_, _ = collection.DeleteOne(context.TODO(), bson.D{
			{
				Key:   "checker",
				Value: checker,
			},
		})
	}

	// 写入新的代码数据
	_, err = collection.InsertOne(context.TODO(), value)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (mongo DataBase) GetTempData(dtype string, checker interface{}, clean bool) (interface{}, error) {

	db := mongo.client.Database(DATABASENAME)
	collection := db.Collection("TempData")

	var temp bson.D

	res := collection.FindOne(context.TODO(), bson.D{
		{
			Key:   "type",
			Value: dtype,
		},
		{
			Key:   "checker",
			Value: checker,
		},
	}).Decode(&temp)

	dmap := temp.Map()

	if res != nil {
		return nil, res
	}

	if clean {
		_, err := collection.DeleteOne(context.TODO(), bson.D{
			{
				Key:   "type",
				Value: dtype,
			},
			{
				Key:   "checker",
				Value: checker,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return dmap["data"], nil
}

func (mongo DataBase) CheckToken(token string, operation string, clean bool) (interface{}, error) {
	db := mongo.client.Database(DATABASENAME)
	collection := db.Collection("Token")

	var temp TokenStruct
	err := collection.FindOne(
		context.TODO(),
		bson.D{
			{
				Key:   "token",
				Value: token,
			},
			{
				Key:   "operation",
				Value: operation,
			},
		},
	).Decode(&temp)

	if err != nil {
		return nil, errors.New("数据不存在")
	}

	if clean {
		_, err := collection.DeleteOne(context.TODO(), bson.D{
			{
				Key:   "token",
				Value: token,
			},
			{
				Key:   "operation",
				Value: operation,
			},
		})
		if err != nil {
			return nil, errors.New("删除 Token 失败")
		}
	}

	// 读取成功，返回 Target ObjectID
	return temp.Target, nil
}

func (mongo DataBase) SaveToken(token string, uid interface{}, operation string) (bool, error) {

	db := mongo.client.Database(DATABASENAME)

	collection := db.Collection("Token")

	tokenValue := TokenStruct{
		Token:     token,
		Target:    uid,
		Operation: operation,
	}

	_, err := collection.InsertOne(context.TODO(), tokenValue)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (mongo DataBase) Register(user UserInfo) (bool, error) {

	db := mongo.client.Database(DATABASENAME)

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

	// // 生成随机密码（用于MCSM）
	// var mcsmPwd map[string]string = make(map[string]string)
	// for _, value := range mongo.config.MCSMConnect {
	// 	mcsmPwd[value.Name] = RandValue(14)
	// }

	// // 预先把已经生成好的 `随机密码` 装入数据库，以免后续还需要进行更新
	// // 大概流程为：
	// // 1. 生成随机密码
	// // 2. 插入到数据库
	// // 3. 使用随机密码批量注册 MCSM 账号
	// // MCSM账号将使用用户的唯一 `ObjectID` 即为 MongoDB 唯一ID
	// user.Mcsmpwd = mcsmPwd

	if err != nil {

		// 插入账号待验证信息
		if mongo.config.EmailConfig.Server != "" && user.Level <= 0 {

			// 不为空则说明配置了邮箱系统信息
			// 自动检测是否支持

			mail := gomail.NewMessage()

			mail.SetHeader("From", mongo.config.EmailConfig.Username)
			mail.SetHeader("To", user.Email)
			mail.SetHeader("Subject", "账号邮箱验证「 "+mongo.config.SiteName+" 」")

			// 加载相应的数据模板文件
			templ, err := mongo.packer.FindString("email/check-email.tmpl")
			if err != nil {
				// 这种错误存在就会不断触发，所以干脆直接崩掉程序
				fmt.Println("Email发送模板不存在。")
				os.Exit(0)
			}

			token := RandStringBytesRmndr(25)

			templ = strings.ReplaceAll(templ, "{site}", mongo.config.SiteName)
			templ = strings.ReplaceAll(templ, "{type}", "register")
			templ = strings.ReplaceAll(templ, "{function}", "注册")
			templ = strings.ReplaceAll(templ, "{domain}", mongo.config.Domain)
			templ = strings.ReplaceAll(templ, "{token}", token)

			mail.SetBody("text/html", templ)

			status, err := SendEmail(mongo.config.EmailConfig, mail)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(0)
			}

			// 邮件发送状态
			if status {

				// 新信息可以插入
				res, err := collection.InsertOne(context.TODO(), user)
				// BulkRegisterMcsmUser(mongo.config, GetObjectID(res.InsertedID), mcsmPwd)
				if err != nil {
					return false, errors.New("数据插入失败")
				}

				_, err = mongo.SaveToken(token, res.InsertedID, "register")
				if err != nil {
					fmt.Println("Token 保存失败：" + token)
					os.Exit(0)
				}

				// 这里的 True 代表开启了 Verify
				return true, nil
			}

		} else {
			// 不需要验证，直接插入
			// 直接将等级更新为 1 不需要进行激活账号
			if user.Level < 1 {
				user.Level = 1
			}
			_, err := collection.InsertOne(context.TODO(), user)
			// BulkRegisterMcsmUser(mongo.config, GetObjectID(res.InsertedID), mcsmPwd)
			if err != nil {
				return false, errors.New("数据插入失败")
			}

			return false, nil
		}
	}

	return false, errors.New("相关数据账号已被注册")
}

func (mongo DataBase) Login(email string, password string) (UserInfo, error) {

	user, err := mongo.GetUser(email)
	if err != nil {
		return user, errors.New("用户信息不存在")
	}

	metaPassword := MetaPassword(password, user.Salt)
	if metaPassword == user.Password {
		// 密码正确
		return user, nil
	}

	return UserInfo{}, errors.New("用户密码不正确")
}

// 这个函数用于用户检查
func (mongo DataBase) GetUser(email string) (UserInfo, error) {

	db := mongo.client.Database(DATABASENAME)
	collection := db.Collection("Users")

	var user UserInfo
	err := collection.FindOne(context.TODO(), bson.D{
		{
			Key:   "email",
			Value: email,
		},
	}).Decode(&user)

	return user, err
}

// 这个函数用于用户更新
func (mongo DataBase) UpdateUser(data bson.D, oriEmail string) error {

	db := mongo.client.Database(DATABASENAME)
	collection := db.Collection("Users")

	_, err := collection.UpdateOne(context.TODO(), bson.D{
		{
			Key:   "email",
			Value: oriEmail,
		},
	}, data)

	return err

}

func (mongo DataBase) AccountLevel(to int, id string) error {

	// 状态码
	// 0 未激活
	db := mongo.client.Database(DATABASENAME)
	collection := db.Collection("Users")

	objectID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{
		{
			Key:   "_id",
			Value: objectID,
		},
	}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "level",
					Value: to,
				},
			},
		},
	}

	// 尝试更新用户状态码
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return errors.New("更新状态失败")
	}

	return nil
}

func (mongo DataBase) Paging(collectName string, onePage int, nowPage int, res interface{}) error {

	skipTo := onePage * nowPage

	db := mongo.client.Database(DATABASENAME)
	collection := db.Collection(collectName)

	cur, err := collection.Find(
		context.TODO(),
		bson.D{},
		options.Find().SetLimit(int64(onePage)),
		options.Find().SetSkip(int64(skipTo)),
	)
	if err != nil {
		return err
	}
	defer cur.Close(context.TODO())

	cur.All(context.TODO(), res)
	return nil
}

func (mongo DataBase) Title() string {
	return mongo.config.SiteName
}

func (mongo DataBase) Config() GeneralConfig {
	return mongo.config
}

func (mongo DataBase) Packer() *packr.Box {
	return mongo.packer
}

func (mongo DataBase) Object() *mongo.Database {
	return mongo.client.Database(DATABASENAME)
}

func GetObjectID(result interface{}) string {
	if oid, ok := result.(primitive.ObjectID); ok {
		return oid.Hex()
	}
	return ""
}

func MetaPassword(password string, salt string) string {
	h := md5.New()
	h.Write([]byte(password + "#" + salt))
	return hex.EncodeToString(h.Sum(nil))
}

func MakePassword(password string) (string, string) {

	randBytes := make([]byte, 10/2)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", ""
	}

	salt := fmt.Sprintf("%x", randBytes)

	return salt, MetaPassword(password, salt)
}

func RandValue(len int) string {
	randBytes := make([]byte, len/2)
	_, err := rand.Read(randBytes)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", randBytes)
}
