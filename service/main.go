// Web Service [main.go]
// Author: mrxiaozhuox <mrxzx.info@gmail.com>
// Date: 2021-10-24
// @FkyOS-MCP

package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitServer() {

	service := gin.Default()

	// 尝试加载 template 目录下的所有页面模板文件
	service.LoadHTMLGlob("template/*")

	service.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})

	service.Run()
}
