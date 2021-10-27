package main

import (
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"fkyos.com/mcp/library"
	"fkyos.com/mcp/service"
	"github.com/gin-gonic/gin"
)

func main() {

	ex, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}
	rootPath := filepath.Dir(ex)

	config, err := library.InitConfig()
	if err != nil {
		log.Fatalln(err)
		os.Exit(0)
	}

	f, err := os.Create(path.Join(rootPath, "log", "service.log"))
	if err != nil {
		panic(err)
	}

	mongo, err := library.MongoConnect(config)
	fatalError(err)
	log.Println("数据库连接测试成功！[PONG]")

	gin.DefaultWriter = io.MultiWriter(f)
	r := gin.Default()

	service.InitServer(r, mongo)
}

func fatalError(err error) {
	if err != nil {
		log.Fatalln(err)
		os.Exit(0)
	}
}
