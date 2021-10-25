package library

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	// _ "github.com/mattn/go-sqlite3"
)

type GeneralConfig struct {
	Hostname       string
	Port           int
	TLS            TLS
	SiteName       string
	MCSMConnect    []MCSMConnect
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

	ex, err := os.Executable()
	if err != nil {
		return GeneralConfig{}, nil
	}

	exPath := filepath.Dir(ex)

	configPath := path.Join(exPath, "config")

	if !exists(configPath) {

		os.MkdirAll(configPath, 0777)

		// 加载配置文件
		general := GeneralConfig{
			Hostname: "0.0.0.0",
			Port:     8848,
			TLS: TLS{
				PemFile: "",
				KeyFile: "",
			},
			MCSMConnect: []MCSMConnect{},
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

	}

	logPath := path.Join(exPath, "log")
	if !exists(logPath) {
		os.MkdirAll(logPath, 0777)
	}

	dbPath := path.Join(exPath, "db")
	if !exists(dbPath) {

		os.MkdirAll(dbPath, 0777)

		// 初始化数据库系统
		// _, err := sql.Open("sqlite3", path.Join(dbPath, "system.db"))
		// if err != nil {
		// 	panic(err)
		// }

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
