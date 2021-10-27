// Web Service [main.go]
// Author: mrxiaozhuox <mrxzx.info@gmail.com>
// Date: 2021-10-24
// @FkyOS-MCP

package service

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"net/http"

	"fkyos.com/mcp/library"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/gin"
)

func InitServer(service *gin.Engine, mongo *library.DataBase) {

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
		apiService(c, mongo)
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

	service.Run()
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

		randBytes := make([]byte, 18/2)
		rand.Read(randBytes)

		salt := string(randBytes)

		password = fmt.Sprintf("%x", md5.Sum([]byte(password)))

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

		_, err := mongo.Register(user)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

}
