package library

import (
	"os"
	"path"
	"path/filepath"

	"gopkg.in/gomail.v2"
)

func SendEmail(conf EmailConfig, tmpl string) (bool, error) {

	dia := gomail.NewDialer(conf.Server, conf.Port, conf.Username, conf.Password)

	ex, err := os.Executable()
	if err != nil {
		return false, err
	}
	rootPath := filepath.Dir(ex)
	templatePath := path.Join(rootPath, "template", "email", tmpl+".tmpl")

	f, err := os.Open(templatePath)
	if err != nil {
		return false, err
	}

	var buffer []byte

	_, err = f.Read(buffer)
	if err != nil {
		return false, err
	}

}
