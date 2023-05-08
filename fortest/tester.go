package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var mySecret = []byte("dick")

type myCustomClaims struct {
	jwt.Claims
	userId   int    `json:"userId"`
	usermail string `json:"usermail"`
}

func main() {
	a, _ := GenerateJwtToken(1, "asdsad")
	ParseJwt(a)

}
func GenerateJwtToken(id int, usermail string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   id,
		"usermail": usermail,
		"nbf":      time.Now().Add(time.Hour * 2).Unix(),
	})

	log.Println(token)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(mySecret)

	return tokenString, err

}

func ParseJwt(token string) (string, error) {
	tokens, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return mySecret, nil

	})
	log.Println(tokens.Claims.(jwt.MapClaims)["usermail"])

	return "", nil

}
