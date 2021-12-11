// Web Service [main.go]
// Author: mrxiaozhuox <mrxzx.info@gmail.com>
// Date: 2021-10-24
// @FkyOS-MCP

package service

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gobuffalo/packr/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/gomail.v2"
	"mrxzx.info/mcp/library"

	"github.com/gin-gonic/gin"
)

func InitServer(service *gin.Engine, db *library.DataBase, config library.GeneralConfig, pack *packr.Box) {

	store := cookie.NewStore([]byte("secret"))
	service.Use(sessions.Sessions("fkyos", store))

	// 写入系统运行时的信息到 TempData 中，用于记录系统运行信息
	_, err := db.SetTempData("system", bson.D{
		{
			Key:   "start-time",
			Value: time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			Key:   "loaded-config",
			Value: config,
		},
	}, "SERVER-INFO")
	// db.GetTempData("system", "SERVER-INFO", false)

	if err != nil {
		log.Fatal(err.Error())
	}

	uint := false
	for _, value := range os.Args {
		if strings.ToUpper(value) == "UINT" {
			uint = true
		}
	}

	// 尝试加载 template 目录下的所有页面模板文件
	if uint {
		service.LoadHTMLGlob("template/**/*")
	} else {
		templates, err := loadTemplate(pack)
		if err != nil {
			panic(err)
		}
		service.SetHTMLTemplate(templates)
	}

	service.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "account/login.tmpl", gin.H{
			"Title": db.Title(),
		})
	})

	service.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "account/register.tmpl", gin.H{
			"Title": db.Title(),
		})
	})

	// Api 接口函数定义
	service.POST("/api/:operation", func(c *gin.Context) {
		apiService(c, db)
	})

	service.GET("/auth", func(c *gin.Context) {
		authType := c.DefaultQuery("type", "register")
		authToken := c.DefaultQuery("token", "")

		if authToken == "" {
			c.HTML(200, "account/auth.tmpl", gin.H{
				"Title":  db.Title(),
				"Result": "Token 信息未输入。",
				"Type":   authType,
				"Status": false,
			})
			return
		}

		if authToken == "AUTH-TOKEN-TEST-SUCCESSFUL" {
			// 这是成功页面测试的 Token 代码，没有任何意义，只会返回成功页面！
			c.HTML(200, "account/auth.tmpl", gin.H{
				"Title":  db.Title(),
				"Result": "账号 Auth 验证成功！",
				"Type":   authType,
				"Status": true,
			})
			return
		}

		oid, err := db.CheckToken(authToken, authType, true)
		if err != nil {
			c.HTML(200, "account/auth.tmpl", gin.H{
				"Title":  db.Title(),
				"Result": "未知的 Token 信息。",
				"Type":   authType,
				"Status": false,
			})
			return
		}

		err = db.AccountLevel(1, library.GetObjectID(oid))
		if err == nil {
			c.HTML(200, "account/auth.tmpl", gin.H{
				"Title":  db.Title(),
				"Result": "账号 Auth 验证成功！",
				"Type":   authType,
				"Status": true,
			})
		} else {
			c.HTML(200, "account/auth.tmpl", gin.H{
				"Title":  db.Title(),
				"Result": "账号解锁失败：" + err.Error(),
				"Type":   authType,
				"Status": false,
			})
		}
	})

	// 自动跳转到 Dashboard 主页中
	service.GET("/center", func(c *gin.Context) {
		c.Redirect(302, "/center/dashboard")
	})

	service.GET("/admin/:page", func(c *gin.Context) {

		page := c.Param("page")
		session := sessions.Default(c)

		// 查询不到用户的登陆信息，跳转到登陆页面
		if session.Get("username") == nil {
			c.Redirect(302, "/login")
			return
		}

		email := session.Get("email")

		// 获取用户具体信息
		user, err := db.GetUser(email.(string))
		if err != nil {

			// 删除用户登录信息
			session.Delete("username")
			session.Delete("email")
			session.Save()

			c.Redirect(302, "/login")
			return
		}

		userPage := c.Query("userPage")
		var userPageNum = 0
		if userPage != "" {
			userPageNum, err = strconv.Atoi(userPage)
			if err != nil {
				userPageNum = 0
			}
		}

		var userList []library.UserInfo
		err = db.Paging("Users", 15, userPageNum, &userList)
		if err != nil {
			panic(err)
		}

		var index = 0
		for _, val := range userList {
			userList[index].Mcsmid = library.GetObjectID(val.Id)[0:12]
			index += 1
		}

		if page == "users" {
			c.HTML(http.StatusOK, "admin/users.tmpl", gin.H{
				"Title":         db.Title(),
				"AcAdmin_Users": true,
				"Username":      session.Get("username"),
				"UserInfo":      user,
				"IsAdmin":       user.Level >= 2,

				"UserList": userList,
			})
			return
		}

	})

	service.GET("/center/:page", func(c *gin.Context) {

		page := c.Param("page")

		session := sessions.Default(c)

		// 查询不到用户的登陆信息，跳转到登陆页面
		if session.Get("username") == nil {
			c.Redirect(302, "/login")
			return
		}

		email := session.Get("email")

		// 生成 Image Hash 值
		h := md5.New()
		h.Write([]byte(email.(string)))
		imageHash := hex.EncodeToString(h.Sum(nil))

		// 获取用户具体信息
		user, err := db.GetUser(email.(string))
		if err != nil {

			// 删除用户登录信息
			session.Delete("username")
			session.Delete("email")
			session.Save()

			c.Redirect(302, "/login")
			return
		}

		if page == "dashboard" {

			// DashBoard 页面操作

			c.HTML(http.StatusOK, "center/dashboard.tmpl", gin.H{
				"Title":       db.Title(),
				"AcDashboard": true,
				"Username":    session.Get("username"),
				"UserInfo":    user,
				"IsAdmin":     user.Level >= 2,

				/* 页面内容数据 */
				"ImageHash":          imageHash,
				"OnlineServerNumber": 0,
				"MSID":               library.GetObjectID(user.Id)[0:12],
			})
			return

		} else if page == "initacc" {

			// log.Println(user)

			// InitAcc 系统信息
			c.HTML(http.StatusOK, "center/initacc.tmpl", gin.H{
				"Title":    db.Title(),
				"Username": session.Get("username"),
				"UserInfo": user,
				"IsAdmin":  user.Level >= 2,
			})
			return

		}

	})

	err = service.Run(config.Hostname + ":" + strconv.Itoa(config.Port))
	if err != nil {
		log.Fatalln(err)
	}
}

// 接口相关服务函数
func apiService(c *gin.Context, mongo *library.DataBase) {

	operation := c.Param("operation")
	session := sessions.Default(c)

	if operation == "register" {

		// 检查必填属性是否存在
		username := c.PostForm("username")
		email := c.PostForm("email")
		password := c.PostForm("password")

		if username == "" || email == "" || password == "" {
			c.JSON(400, gin.H{
				"error": "缺少必须参数",
			})
			return
		}

		randBytes := make([]byte, 10/2)
		_, err := rand.Read(randBytes)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "系统运行错误",
			})
			return
		}

		salt := fmt.Sprintf("%x", randBytes)

		h := md5.New()
		h.Write([]byte(password))

		password = hex.EncodeToString(h.Sum(nil))

		// 初始化的用户信息
		// status 0 代表用户未在线
		// level  0 代表用户未激活
		user := library.UserInfo{
			Username: username,
			Password: password,
			Salt:     salt,
			Email:    email,
			About:    "Hello World!",
			Status:   0,
			Level:    0,
			Regtime:  primitive.NewDateTimeFromTime(time.Now()),
			Initacc:  false,
		}

		needVerify, err := mongo.Register(user)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		} else {
			c.JSON(200, gin.H{
				"status": "成功",
				"verify": needVerify,
			})
			return
		}

	} else if operation == "login" {
		// 登录操作

		email := c.PostForm("email")
		password := c.PostForm("password")

		if email == "" || password == "" {
			c.JSON(400, gin.H{
				"error": "缺少必须参数",
			})
			return
		}

		user, err := mongo.Login(email, password)
		if err != nil {
			// fmt.Println(err)
			c.JSON(400, gin.H{
				"error": "用户登录失败",
			})
			return
		}

		session.Set("username", user.Username)
		session.Set("email", user.Email)
		err = session.Save()
		if err != nil {
			c.JSON(400, gin.H{
				"error": "用户登录失败",
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "成功",
		})
		return
	} else if operation == "logout" {

		if session.Get("username") != nil {

			session.Delete("username")
			session.Delete("email")
			_ = session.Save()
		}

		c.JSON(200, gin.H{
			"status": "成功",
		})
		return

	} else if operation == "edit-info" {

		username := session.Get("username")
		email := session.Get("email")

		if username == nil || email == nil {
			c.JSON(401, gin.H{
				"error": "用户未登录",
			})
			return
		}

		user, err := mongo.GetUser(email.(string))
		if err != nil {
			c.JSON(500, gin.H{
				"error": "获取信息失败",
			})
			return
		}

		newUsername := c.PostForm("username")
		newPassword := c.PostForm("password")

		var newValue bson.D
		if newUsername != "" && newPassword != "" {
			newValue = bson.D{
				{
					Key:   "username",
					Value: newUsername,
				},
				{
					Key:   "password",
					Value: library.MetaPassword(newPassword, user.Salt),
				},
			}
		} else if newUsername != "" {
			newValue = bson.D{
				{
					Key:   "username",
					Value: newUsername,
				},
			}
		} else {
			newValue = bson.D{
				{
					Key:   "password",
					Value: newPassword,
				},
			}
		}

		// 编辑用户个人信息
		config := mongo.Config()
		if (config.EmailConfig != library.EmailConfig{}) {

			mail := gomail.NewMessage()

			mail.SetHeader("From", mongo.Config().EmailConfig.Username)
			mail.SetHeader("To", user.Email)
			mail.SetHeader("Subject", "账号邮箱验证「 "+mongo.Config().SiteName+" 」")

			// 加载相应的数据模板文件
			templ, err := mongo.Packer().FindString("email/send-token.tmpl")
			if err != nil {
				// 这种错误存在就会不断触发，所以干脆直接崩掉程序
				fmt.Println("Email发送模板不存在。")
				os.Exit(0)
			}

			token := library.RandStringBytesRmndr(10)

			templ = strings.ReplaceAll(templ, "{site}", mongo.Config().SiteName)
			templ = strings.ReplaceAll(templ, "{type}", "edit")
			templ = strings.ReplaceAll(templ, "{function}", "账号编辑")
			templ = strings.ReplaceAll(templ, "{token}", token)

			mail.SetBody("text/html", templ)

			_, err = library.SendEmail(config.EmailConfig, mail)
			if err != nil {
				panic(err)
			}

			oldToken, err := mongo.GetTempData("account-token", email.(string)+".eda", true)
			if err == nil {
				dataMap := oldToken.(primitive.D)
				// fmt.Println(dataMap.Map()["token"])
				_, _ = mongo.GetTempData("token-check", dataMap.Map()["token"], true)
			}

			_, err = mongo.SetTempData("token-check", bson.D{
				{
					Key:   "operation",
					Value: "update",
				},
				{
					Key:   "target",
					Value: "Users",
				},
				{
					Key: "filter",
					Value: bson.D{
						{
							Key:   "email",
							Value: email,
						},
					},
				},
				{
					Key: "value",
					Value: bson.D{
						{
							Key:   "$set",
							Value: newValue,
						},
					},
				},
			}, token)
			if err != nil {
				c.JSON(500, gin.H{
					"error": "数据储存失败",
				})
				return
			}

			mongo.SetTempData("account-token", bson.D{
				{
					Key:   "token",
					Value: token,
				},
			}, email.(string)+".eda")

		} else {

			err = mongo.UpdateUser(newValue, email.(string))
			if err != nil {
				c.JSON(500, gin.H{
					"error": "数据储存失败",
				})
				return
			}

		}

		c.JSON(200, gin.H{
			"status": "成功",
		})
		return

	} else if operation == "initacc" {

		// 原始数据（因为这里涉及到数据更新，所以说使用 oriX 表示）
		oriUsername := session.Get("username")
		oriEmail := session.Get("email")

		newUsername := c.PostForm("username")
		newEmail := c.PostForm("email")
		newPassword := c.PostForm("password")

		if oriEmail == nil || oriUsername == nil {
			c.JSON(401, gin.H{
				"error": "用户未登录",
			})
			return
		}

		oriUser, err := mongo.GetUser(oriEmail.(string))
		if err != nil {
			c.JSON(500, gin.H{
				"error": "用户信息错误",
			})
			return
		}

		if !oriUser.Initacc {
			c.JSON(400, gin.H{
				"error": "账号无权使用本服务",
			})
			return
		}

		_, state := mongo.GetUser(newEmail)
		if state == nil {
			c.JSON(400, gin.H{
				"error": "邮箱信息已存在",
			})
			return
		}

		newSalt, newMetaPassword := library.MakePassword(newPassword)

		err = mongo.UpdateUser(bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{
						Key:   "email",
						Value: newEmail,
					},
					{
						Key:   "username",
						Value: newUsername,
					},
					{
						Key:   "password",
						Value: newMetaPassword,
					},
					{
						Key:   "salt",
						Value: newSalt,
					},
					{
						Key:   "initacc",
						Value: false,
					},
				},
			},
		}, oriUser.Email)

		if err != nil {
			c.JSON(500, gin.H{
				"error": "服务器数据插入失败",
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "成功",
		})
		return

	} else if operation == "mcsm-load" {

		// 重新对账号进行 MCSM 账号注册！

		// 获取当前登录的用户信息
		// username := session.Get("username")
		email := session.Get("email")

		user, err := mongo.GetUser(email.(string))

		if err != nil {
			c.JSON(500, gin.H{
				"error": "用户信息错误",
			})
			return
		}

		var Mcsmpwd map[string]string = make(map[string]string)
		if user.Mcsmpwd != nil {
			Mcsmpwd = user.Mcsmpwd
		}

		var count int = 0
		for _, conn := range mongo.Config().MCSMConnect {

			if _, ok := user.Mcsmpwd[conn.Name]; ok {
				// 已经存在了，则往后继续查找
				_ = user.Mcsmpwd[conn.Name]
				continue
			}

			// 重新为没有注册过的账号进行注册
			temp := library.RandValue(14)
			err := library.RegisterMcsmUser(library.GetObjectID(user.Id), temp, conn.Domain, conn.MasterToken)
			if err == nil {
				count += 1
				Mcsmpwd[conn.Name] = temp
			}
		}

		err = mongo.UpdateUser(bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{
						Key:   "mcsmpwd",
						Value: Mcsmpwd,
					},
				},
			},
		}, user.Email)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "信息更新错误",
			})
			return
		}

		// 将成功的次数返回到接口中
		c.JSON(200, gin.H{
			"status": count,
		})
		return

	} else if operation == "token-check" {

		username := session.Get("username")
		email := session.Get("email")

		if username == nil || email == nil {
			c.JSON(401, gin.H{
				"error": "用户未登录",
			})
			return
		}

		token := c.PostForm("token")
		if token == "" {
			c.JSON(400, gin.H{
				"error": "参数不足",
			})
			return
		}

		value, err := mongo.GetTempData("token-check", token, true)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "数据获取错误",
			})
			return
		}
		if value == nil {
			c.JSON(400, gin.H{
				"error": "Token不存在",
			})
			return
		}
		valueMap := value.(bson.D).Map()

		if valueMap["operation"] == "update" {
			updateTarget := valueMap["target"]
			collection := mongo.Object().Collection(updateTarget.(string))
			_, err = collection.UpdateOne(context.TODO(), valueMap["filter"].(bson.D), valueMap["value"].(bson.D))
			if err != nil {
				c.JSON(500, gin.H{
					"error": "数据更新错误",
				})
				return
			}
		}

		// 清除登录状态
		session.Delete("username")
		session.Delete("email")
		_ = session.Save()

		c.JSON(200, gin.H{
			"status": "成功",
		})
		return
	}
}

func loadTemplate(box *packr.Box) (*template.Template, error) {

	t := template.New("")

	for _, path := range box.List() {

		if path == ".DS_Store" {
			continue
		}

		s, err := box.FindString(path)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		t, err = t.New(path).Parse(s)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		fmt.Println("Loading template: ", path)
	}
	return t, nil
}
