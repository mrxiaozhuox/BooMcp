package main

import (
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
	"mrxzx.info/mcp/library"
	"mrxzx.info/mcp/service"
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
	if !exists(path.Join(rootPath, "config", "server.ini")) {
		err = mongo.InitDataBase()
		fatalError(err)

		// 创建服务器已初始化标记
		_, err = os.Create(path.Join(rootPath, "config", "server.ini"))
		fatalError(err)

		log.Println("数据库首次初始化成功！[MONGO]")
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
