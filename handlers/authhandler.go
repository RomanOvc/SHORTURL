package handlers

import (
	"appurl/repository"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type AuthInquirysRepository struct {
	Psql  *repository.AuthInquirysRepository
	Redis *repository.RedisClient
}

func NewAuthInquirysRepository(postgres *repository.AuthInquirysRepository, redis *repository.RedisClient) *AuthInquirysRepository {
	return &AuthInquirysRepository{Psql: postgres, Redis: redis}
}

type UserInfoStruct struct {
	UserId   int    `json:"userId"`
	Usermail string `json:"usermail"`
	Password string `json:"password"`
}

func (rep *AuthInquirysRepository) CreateUserH(w http.ResponseWriter, r *http.Request) {
	var (
		userInfo      UserInfoStruct
		handlerResult []byte
		err           error
	)

	defer func() {
		if err != nil {
			log.Println(err, "Error request")
			w.WriteHeader(400)
			w.Write(nil)
		} else {
			w.Write(handlerResult)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	err = json.NewDecoder(r.Body).Decode(&userInfo)
	if err != nil {
		return
	}
	//  если хотя бы одно поле пришло пустым возвращать 400
	// FIXME UserEmail
	if userInfo.Usermail == "" || userInfo.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// иначе проверяем, есть ли пользователь с таким mail в базе
	emptyUser := rep.Psql.SelectUserMail(r.Context(), userInfo.Usermail)

	// если нет пользователя добавлем его в базу и отправлем письмо, условныи оператор - костыль, надо переписать
	if emptyUser == 0 {

		hashPassword, _ := HashPassword(userInfo.Password)

		generateActivUuid, err := rep.Psql.CreateUser(r.Context(), userInfo.Usermail, hashPassword, false)
		if err != nil {
			return
		}

		// путь убрать в .env
		strUrl := "http://127.0.0.1:8000/create_user/activate/" + generateActivUuid
		_, err = Gsmtp(userInfo.Usermail, strUrl)
		if err != nil {
			return
		}
	}
}

// переход по ссылке из письма
func (rep *AuthInquirysRepository) EmailActivateH(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	index := vars["uuid"]

	err := rep.Psql.CheckUserUuidToEmail(r.Context(), index)
	if err != nil {
		// FIXME handle this error
		return
	}
}

// 1) Проверить наличие пользователя в базе, правильность введеных логина и пороля
// 2) Если 1 этап успешен проверить поле activate если true - выдать токен, если не активен отправить письмо на почту, удалив предыдущу запись в emailactivate
// 3) миделваром обернуть остальные эндпоинты
// {
// "usermail":"you mail",
// "password":"pass"
// }
func (rep *AuthInquirysRepository) AuthentificateUserH(w http.ResponseWriter, r *http.Request) {
	// как обработать ошибку, если тель запроса пустое вовсе
	var (
		userInfo UserInfoStruct
		u        []byte
		err      error
	)

	defer func() {
		if err != nil {
			log.Println(err, "Error request")
			w.WriteHeader(400)
			w.Write(nil)
		} else {
			w.Write(u)
		}
	}()
	w.Header().Set("Content-Type", "application/json")

	err = json.NewDecoder(r.Body).Decode(&userInfo)
	if err != nil {
		err = errors.Wrap(err, "decode")
		return
	}
	log.Println(userInfo)

	user := rep.Psql.SelectUser(r.Context(), userInfo.Usermail)

	if user.Usermail == "" || r.Body == http.NoBody {
		// если пользователя нет, то статус 400
		log.Println("bad user")
		w.WriteHeader(http.StatusBadRequest)
		// FIXME добавь return и сотри else
	} else {
		// проверка пароля
		checkPass := bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(userInfo.Password))
		if checkPass == nil {
			// генерация access и refresh токена
			var (
				updateFieldRefreshToken int
				addAccessTokenToRedis   string
			)

			accessToken, _ := GenerateAcceessToken(user.UserId, user.Usermail)
			refreshToken, _ := GenerateRefreshToken(user.UserId)

			// добавить токен в редис
			addAccessTokenToRedis, err = rep.Redis.AddAccessToken(r.Context(), user.Usermail, accessToken)
			if err != nil {
				return
			}
			log.Println(addAccessTokenToRedis)
			// обнавление refresh token в db
			updateFieldRefreshToken, err = rep.Psql.UpdateRefershTokenForUser(r.Context(), user.UserId, refreshToken)
			if err != nil {
				return
			}
			log.Println("updateFieldRefreshToken ", updateFieldRefreshToken)

			if updateFieldRefreshToken != 0 {
				u, err = json.Marshal(AccessRefreshToken(accessToken, refreshToken))
				if err != nil {
					// FIXME handle error
					err = errors.Wrap(err, "marshal")
					return
				}
				log.Println("acceess and refresh tokens access")
			}

		} else {
			// иначе статус 400
			w.WriteHeader(http.StatusBadRequest)
			log.Println("bad pass")
		}
	}
}

type RefreshTokenStruct struct {
	RefreshToken string `json:"refresh_token"`
}
type MessageReq struct {
	Message string
}

func (rep *AuthInquirysRepository) RefreshTokenH(w http.ResponseWriter, r *http.Request) {
	// 1) получаем refresh token  формате json +
	// 2) парсим токен +
	// 3) получаем claims["user_id"] +
	// 4) Делаем запрос в базу,
	// 5) проверяем наличие user_id.
	// 6) если такой есть проверяем срок годности refresh токена и валидность
	// 7) если refresh token е протух, то гененрируем новую пару токенов
	// 8) если refresh token протух, перенаправляем на auth
	var (
		u   []byte
		err error
	)
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err != nil {
			log.Println(err, "Error request")
			w.WriteHeader(400)
			w.Write(nil)
		} else {
			w.Write(u)
		}
	}()
	var rT RefreshTokenStruct
	err = json.NewDecoder(r.Body).Decode(&rT)
	if err != nil {
		return
	}

	token, err := jwt.Parse(rT.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrap(err, "token not parsed")
		}
		return mySignedAccessRefreshToken, err
	})
	//

	userId := int(token.Claims.(jwt.MapClaims)["userId"].(float64))
	tokenExp := int(token.Claims.(jwt.MapClaims)["exp"].(float64))
	timeNow := int(time.Now().Unix())

	// проверка наличия пользователя по id
	checkUserId := rep.Psql.SelectByUserId(r.Context(), userId)
	// FIXME
	// if checkUserId.UserId == 0 || tokenExp > timeNow || rT.RefreshToken != checkUserId.RefreshToken {
	// ERROR
	// }
	if checkUserId.UserId != 0 && tokenExp > timeNow && rT.RefreshToken == checkUserId.RefreshToken {
		// генерация токенов
		accessToken, _ := GenerateAcceessToken(checkUserId.UserId, checkUserId.Usermail)
		refreshToken, _ := GenerateRefreshToken(checkUserId.UserId)
		// добавление access токена в redis
		addAccessTokenToRedis, err := rep.Redis.AddAccessToken(r.Context(), checkUserId.Usermail, accessToken)
		if err != nil {
			return
		}
		log.Println(addAccessTokenToRedis)

		// обнавление поля refresh токена в postgres
		updateFieldRefreshToken, err := rep.Psql.UpdateRefershTokenForUser(r.Context(), checkUserId.UserId, refreshToken)
		if err != nil {
			return
		}
		log.Println("updateFieldRefreshToken ", updateFieldRefreshToken)

		// FIXME if updateFieldRefreshToken == 0 then error
		if updateFieldRefreshToken != 0 {
			u, err = json.Marshal(AccessRefreshToken(accessToken, refreshToken))
			if err != nil {
				return
			}
			log.Println("acceess and refresh tokens access")
		}
		// если присланныи refresh токен валидныи и не протух, генерим новую пару, иначе отправлеяем пользователя на auth
	} else {
		// FIXME не нужно
		// юсер не валидныи
		log.Println("not valid")
		u, err = json.Marshal(&MessageReq{Message: "http://127.0.0.1:8000/create_user"})
		if err != nil {
			return
		}
	}
}

// FIXME перенеси её в отдельный файл example: "middleware.go"
// middleware for checck token
func (rep *AuthInquirysRepository) IsAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var tokenHeader = r.Header["Token"]

		if r.Header["Token"][0] == "" {
			w.WriteHeader(http.StatusUnauthorized)
		} else {

			// парсим токен
			token, _ := jwt.Parse(tokenHeader[0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("There was an error")
				}
				return mySignedAccessRefreshToken, nil
			})

			// claim по "exp" - expire time
			tokenExp := int(token.Claims.(jwt.MapClaims)["exp"].(float64))
			// текущее время
			timeNow := int(time.Now().Unix())
			// claim по "usermail" - имя пользователя
			usermail := token.Claims.(jwt.MapClaims)["usermail"].(string)
			// проверяем наличие ключа в redis. Если успешно, возвращаем acceess токен по ключу
			accessTokenFromRedis, _ := rep.Redis.GetAccessTokenByUsermail(r.Context(), usermail)

			log.Println("this token", r.Header["Token"][0])
			log.Println("access token redis ", accessTokenFromRedis)

			// проверяется полученыи токен от пользователя на время жизни токена, валидность, и наличие в redis
			if token.Raw == accessTokenFromRedis && tokenExp > timeNow && token.Valid {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(401)
				w.Write(nil)
				return
			}
		}
	})
}

// запрос на смену парол
// пользователь вводит свои пароль
type ChangePassStruct struct {
	Email string `json:"email"`
}

func (rep *AuthInquirysRepository) ForgotPasswordH(w http.ResponseWriter, r *http.Request) {
	log.Println("changepass")
	// если пользователь есть отправить письмо на почту, если нет сообщить что пользователя нет
	var changePass ChangePassStruct
	err := json.NewDecoder(r.Body).Decode(&changePass)
	if err != nil {
		return
	}

	log.Println(changePass.Email)

}
