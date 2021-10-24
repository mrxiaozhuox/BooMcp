// Web Service [main.go]
// Author: mrxiaozhuox <mrxzx.info@gmail.com>
// Date: 2021-10-24
// @FkyOS-MCP

package service

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/gin"
)

func InitServer() {

	service := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	service.Use(sessions.Sessions("fkyos", store))

	// 尝试加载 template 目录下的所有页面模板文件
	service.LoadHTMLGlob("template/**/*")

	service.GET("/center/", func(c *gin.Context) {
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
				"title":              "FkyOS Server",
				"AcDashboard":        true,
				"Username":           "mrxiaozhuox",
				"OnlineServerNumber": 0,
			})
			return
		}
	})

	service.Run()
}
