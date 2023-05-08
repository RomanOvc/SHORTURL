package handlers

import (
	"log"
	"net/smtp"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// FIXME SendEmail
func Gsmtp(usermail, genUrl string) (bool, error) {
	body := genUrl
	from := "roman.ovcharov.1997@gmail.com"
	pass := "wfxkixwblqzwnref"
	to := usermail

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hello there\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("wtf", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return false, errors.Wrap(err, "handler/gsmtp/Gsmtp() error")
	}

	return true, err
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
