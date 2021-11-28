package library

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type GeneralConfig struct {
	Hostname    string
	Port        int
	TLS         TLS
	Domain      string
	SiteName    string
	MCSMConnect []MCSMConnect
	MongoDbURI  string
	EmailConfig EmailConfig
}

type TLS struct {
	PemFile string
	KeyFile string
}

type EmailConfig struct {
	Server   string
	Port     int
	Username string
	Password string
}

type MCSMConnect struct {
	Name          string
	Domain        string
	ApiKey        string
	MasterToken   string
	MaximumMemory int
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
			SiteName:    "BooMcp",
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

			if tmp.Domain == "" {
				// 如果域名为空，则自动生成 IP + 端口 路径

				protocol := "https"
				if (tmp.TLS == TLS{}) {
					protocol = "http"
				}

				host := tmp.Hostname
				if tmp.Hostname == "0.0.0.0" {
					host = GetInternalIP()
					log.Println("服务将运行于内网IP：0.0.0.0 [Internal]")
				}

				url := protocol + "://" + host + ":" + strconv.Itoa(tmp.Port)
				tmp.Domain = url
			}

			log.Println("服务器加载配置成功！[Successful]")
			log.Println("服务器运行地址：" + tmp.Domain)

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

func GetInternalIP() string {
	// 思路来自于Python版本的内网IP获取，其他版本不准确
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	// udp 面向无连接，所以这些东西只在你本地捣鼓
	res := conn.LocalAddr().String()
	res = strings.Split(res, ":")[0]
	return res
}
