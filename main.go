package main

import (
	"fkyos.com/mcp/library"
	"fkyos.com/mcp/service"
	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	library.InitConfig()
	// if err != nil {
	// 	panic(err)
	// }

	service.InitServer(r)
}
