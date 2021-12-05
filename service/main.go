// Web Service [main.go]
// Author: mrxiaozhuox <mrxzx.info@gmail.com>
// Date: 2021-10-24
// @FkyOS-MCP

package service

import (
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
			})
			return

		} else if page == "initacc" {

			log.Println(user)

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

		// 编辑用户个人信息

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

		for _, conn := range mongo.Config().MCSMConnect {

			if _, ok := user.Mcsmpwd[conn.Name]; ok {
				// 已经存在了，则往后继续查找
				_ = user.Mcsmpwd[conn.Name]
				continue
			}

		}

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
