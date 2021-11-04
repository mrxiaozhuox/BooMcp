package main

import (
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"

	"fkyos.com/mcp/library"
	"fkyos.com/mcp/service"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
)

func main() {

	rand.Seed(time.Now().UnixNano())

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

	pack := packr.NewBox("./template")

	f, err := os.Create(path.Join(rootPath, "log", "service.log"))
	if err != nil {
		panic(err)
	}

	mongo, err := library.MongoConnect(config, pack)
	fatalError(err)
	log.Println("数据库连接测试成功！[PONG]")

	gin.DefaultWriter = io.MultiWriter(f)
	r := gin.Default()

	service.InitServer(r, mongo, config)
}

func fatalError(err error) {
	if err != nil {
		log.Fatalln(err)
		os.Exit(0)
	}
}
