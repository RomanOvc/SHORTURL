package middlewars

import (
	"appurl/handlers"
	"appurl/repository"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authInquirysRepository struct {
	Psql  *repository.AuthInquirysRepository
	Redis *repository.RedisClient
}

func NewAuthInquirysRepository(postgres *repository.AuthInquirysRepository, redis *repository.RedisClient) *authInquirysRepository {
	return &authInquirysRepository{Psql: postgres, Redis: redis}
}

func IsAuth(postgres *repository.AuthInquirysRepository, redis *repository.RedisClient) func(next http.HandlerFunc) http.Handler {
	return func(next http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenHeader = r.Header["Token"]

			defer r.Body.Close()

			w.Header().Set("Content-Type", "application/json")

			if r.Header["Token"][0] == "" {
				w.WriteHeader(http.StatusUnauthorized)
			} else {

				// парсим токен
				token, _ := jwt.Parse(tokenHeader[0], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("there was an error")
					}
					return handlers.Secret, nil
				})

				// claim по "exp" - expire time
				tokenExp := int(token.Claims.(jwt.MapClaims)["exp"].(float64))
				// текущее время
				timeNow := int(time.Now().Unix())
				// claim по "usermail" - имя пользователя
				usermail := token.Claims.(jwt.MapClaims)["usermail"].(string)
				activate := token.Claims.(jwt.MapClaims)["activate"].(bool)

				// проверяем наличие ключа в redis. Если успешно, возвращаем acceess токен по ключу
				accessTokenFromRedis, err := redis.GetAccessTokenByUsermail(r.Context(), usermail)
				if err != nil {
					log.Printf("GetAccessTokenByUsermail(): %s", err.Error())
				}

				// проверяется полученыи токен от пользователя на время жизни токена, валидность, и наличие в redis
				if token.Raw == accessTokenFromRedis && tokenExp > timeNow && token.Valid {
					if !activate {
						w.WriteHeader(401)
						w.Write([]byte(`{"message":"account not active"}`))
					} else {
						next.ServeHTTP(w, r)
					}
				} else {
					w.WriteHeader(401)
					w.Write([]byte(`{"message":"token is death"}`))

					return
				}
			}
		})
	}
}
