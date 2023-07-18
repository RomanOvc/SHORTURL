package handlers

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func SendEmailToConfirm(userEmail, genUrl string) (bool, error) {
	body := genUrl
	from :="youremailadress"
	pass := "wfxkixwblqzwnref"
	to := userEmail

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hello there\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("wtf", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return false, fmt.Errorf("handler/utils SendEmailToConfirm() error %w", err)
	}

	return true, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", errors.Wrap(err, "error generation pass")
	}

	return string(bytes), nil

}

func ChechHashPass(hashPass string, originPass string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashPass), []byte(originPass))
	if err != nil {
		errors.Wrap(err, "hash pass != origin pass ")
	}

	return err
}

func SendEmailToPassReset(userEmail, resetToken string) error {
	body := "http://127.0.0.1:8001/auth/resetpass/" + resetToken

	from := "roman.ovcharov.1997@gmail.com"
	pass := "wfxkixwblqzwnref"
	to := userEmail

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hello there\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("wtf", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return errors.Wrap(err, "handler/gsmtp/Gsmtp() error")
	}

	return nil
}

func AccessTokenParce(token string) (string, error) {
	tokenString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return Secret, nil
	})
	userEmail := tokenString.Claims.(jwt.MapClaims)["usermail"].(string)

	return userEmail, err
}

func TokenParse(token string) (*jwt.Token, error) {
	tokenString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return Secret, nil
	})
	if err != nil {
		log.Println("error auth")

		return nil, err
	}
	return tokenString, nil
}
