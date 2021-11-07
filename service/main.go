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
	"log"
	"net/http"
	"strconv"

	"fkyos.com/mcp/library"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/gin"
)

func InitServer(service *gin.Engine, db *library.DataBase, config library.GeneralConfig) {

	store := cookie.NewStore([]byte("secret"))
	service.Use(sessions.Sessions("fkyos", store))

	// 尝试加载 template 目录下的所有页面模板文件
	service.LoadHTMLGlob("template/**/*")

	service.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "account/login.tmpl", gin.H{
			"Title": "FkyOS Server",
		})
	})

	service.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "account/register.tmpl", gin.H{
			"Title": "FkyOS Server",
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
		c.Redirect(http.StatusMovedPermanently, "/center/dashboard")
	})

	service.GET("/center/:page", func(c *gin.Context) {

		page := c.Param("page")

		session := sessions.Default(c)

		// 查询不到用户的登陆信息，跳转到登陆页面
		if session.Get("UserID") == nil {
			c.Redirect(http.StatusMovedPermanently, "/login")
			return
		}

		if page == "dashboard" {
			// DashBoard 页面操作
			c.HTML(http.StatusOK, "center/dashboard.tmpl", gin.H{
				"Title":              "FkyOS Server",
				"AcDashboard":        true,
				"Username":           "mrxiaozhuox",
				"OnlineServerNumber": 0,
			})
			return
		}
	})

	err := service.Run(config.Hostname + ":" + strconv.Itoa(config.Port))
	if err != nil {
		log.Fatalln(err)
	}
}

func apiService(c *gin.Context, mongo *library.DataBase) {

	operation := c.Param("operation")

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
			Sex:      2,
			About:    "",
			Status:   0,
			Level:    0,
		}

		_, err = mongo.Register(user)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		} else {
			c.JSON(200, gin.H{
				"status": "successful",
			})
		}

	} else if operation == "login" {

		// 登录操作

	}

}
