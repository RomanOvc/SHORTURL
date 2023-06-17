package handlers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

type AccessAndRefreshToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var Secret = []byte("shortUrlProjectAccessRefreshKey")

func GenerateAcceessToken(userId int, usermail string, activate bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   userId,
		"usermail": usermail,
		"activate": activate,
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
	})

	tokenString, err := token.SignedString(Secret)
	if err != nil {
		errors.Wrap(err, "error SignedString()")
	}

	return tokenString, err
}

func GenerateRefreshToken(userId int, activate bool) (string, error) {

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   userId,
		"activate": activate,
		"exp":      time.Now().Add(time.Minute * 48).Unix(),
	})

	tokenRefreshString, err := refreshToken.SignedString(Secret)
	if err != nil {
		errors.Wrap(err, "SignedString")
	}

	return tokenRefreshString, err
}

func AccessRefreshToken(accessToken, refreshToken string) *AccessAndRefreshToken {
	return &AccessAndRefreshToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
