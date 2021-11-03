package library

import (
	"gopkg.in/gomail.v2"
)

func SendEmail(conf EmailConfig, message *gomail.Message) (bool, error) {

	dia := gomail.NewDialer(conf.Server, conf.Port, conf.Username, conf.Password)

	if err := dia.DialAndSend(message); err != nil {
		return false, err
	}

	return true, nil
}
