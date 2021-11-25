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
	"github.com/gobuffalo/packr/v2"
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

	pack := packr.New("Template", "./template")

	f, err := os.Create(path.Join(rootPath, "log", "service.log"))
	if err != nil {
		panic(err)
	}

	mongo, err := library.MongoConnect(config, pack)
	fatalError(err)
	log.Println("数据库连接测试成功！[PONG]")

	// 如果数据库未被初始化则初始化数据库信息
	if !exists(path.Join(rootPath, "log", ".init_server")) {
		mongo.
	}

	gin.DefaultWriter = io.MultiWriter(f)
	r := gin.Default()

	service.InitServer(r, mongo, config, pack)
}

func fatalError(err error) {
	if err != nil {
		log.Fatalln(err)
		os.Exit(0)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func isFile(path string) bool {
	return !isDir(path)
}
