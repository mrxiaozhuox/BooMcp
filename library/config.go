package library

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

type GeneralConfig struct {
	Hostname       string
	Port           int
	TLS            TLS
	SiteName       string
	MCSMConnect    []MCSMConnect
	MongoDbURI     string
	RegisterConfig RegisterConfig
}

type TLS struct {
	PemFile string
	KeyFile string
}

type RegisterConfig struct {
	AllowRegister bool
}

type MCSMConnect struct {
	Domain     string
	ApiKey     string
	Active     bool
	UpdateTime int
	checkTime  int
}

func InitConfig() (c GeneralConfig, err error) {

	log.Println("服务器启动中...")

	ex, err := os.Executable()
	if err != nil {
		return GeneralConfig{}, nil
	}

	exPath := filepath.Dir(ex)

	// 日志目录检查并创建
	logPath := path.Join(exPath, "log")
	if !exists(logPath) {
		PanicErr(os.MkdirAll(logPath, 0777))
	}

	configPath := path.Join(exPath, "config")
	if !exists(configPath) {

		PanicErr(os.MkdirAll(configPath, 0777))

		// 加载配置文件
		general := GeneralConfig{
			Hostname: "0.0.0.0",
			Port:     8848,
			TLS: TLS{
				PemFile: "",
				KeyFile: "",
			},
			MCSMConnect: []MCSMConnect{},
			MongoDbURI:  "mongodb://localhost:27017",
			RegisterConfig: RegisterConfig{
				AllowRegister: true,
			},
		}

		jsons, err := json.MarshalIndent(general, "", "    ")
		if err != nil {
			return GeneralConfig{}, err
		}

		err = ioutil.WriteFile(path.Join(configPath, "property.json"), jsons, 0777)
		if err != nil {
			return GeneralConfig{}, err
		}

		log.Println("服务器首次初始化成功！[Successful]")

		return general, nil

	} else {
		// 文件存在，则自动读取并加载
		if isFile(path.Join(configPath, "property.json")) {

			file, err := os.Open(path.Join(configPath, "property.json"))
			PanicErr(err)
			defer file.Close()

			content, err := ioutil.ReadAll(file)
			PanicErr(err)

			var tmp GeneralConfig
			err = json.Unmarshal(content, &tmp)
			PanicErr(err)

			log.Println("服务器加载配置成功！[Successful]")

			return tmp, nil
		}
	}

	return GeneralConfig{}, nil
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

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}
