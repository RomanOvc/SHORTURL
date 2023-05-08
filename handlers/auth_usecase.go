package handlers

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

type AccessAndRefreshToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var mySignedAccessRefreshToken = []byte("shortUrlProjectAccessRefreshKey")

func GenerateAcceessToken(userId int, usermail string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   userId,
		"usermail": usermail,
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
	})

	log.Println(token)
	tokenString, err := token.SignedString(mySignedAccessRefreshToken)
	if err != nil {
		errors.Wrap(err, "error SignedString()")
	}

	return tokenString, err
}

func GenerateRefreshToken(userId int) (string, error) {
	refreshToken := jwt.New(jwt.SigningMethodHS256)

	claims := refreshToken.Claims.(jwt.MapClaims)
	claims["userId"] = userId
	claims["exp"] = time.Now().Add(time.Hour * 48).Unix()

	tokenRefreshString, err := refreshToken.SignedString(mySignedAccessRefreshToken)
	if err != nil {
		errors.Wrap(err, "error SignedString()")
	}
	return tokenRefreshString, err
}

func AccessRefreshToken(accessToken, refreshToken string) *AccessAndRefreshToken {
	return &AccessAndRefreshToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

// запрос на смену пароля
