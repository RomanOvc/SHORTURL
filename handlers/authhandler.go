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

const (
	confirmURL = "http://127.0.0.1:8001/auth/confirm/"
)

type AuthInquirysRepository struct {
	Psql  *repository.AuthInquirysRepository
	Redis *repository.RedisClient
}

func NewAuthInquirysRepository(postgres *repository.AuthInquirysRepository, redis *repository.RedisClient) *AuthInquirysRepository {
	return &AuthInquirysRepository{Psql: postgres, Redis: redis}
}

type UserInfoStruct struct {
	UserId          int    `json:"userId"`
	UserEmail       string `json:"usermail"`
	ActivateAccount bool   `json:"activate"`
	Password        string `json:"password"`
}

type MessageError struct {
	Message string `json:"message"`
}

func (rep *AuthInquirysRepository) CreateUserH(w http.ResponseWriter, r *http.Request) {
	var (
		userInfo    UserInfoStruct
		message     []byte
		err         error
		statusEmail bool
	)

	defer func() {
		if err != nil {
			log.Println(err, "Error request")
			w.WriteHeader(400)
			w.Write(nil)
		} else {
			w.Write(message)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	err = json.NewDecoder(r.Body).Decode(&userInfo)
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "invalid params"})
		log.Println("body is empty")

		return
	}

	// иначе проверяем, есть ли пользователь с таким mail в базе
	emptyUser, err := rep.Psql.SelectUserIdByMail(r.Context(), userInfo.UserEmail)
	if err != nil {
		log.Println("error: handler/authandler CreateUserH() in SelectUserIdByMail() ")
		message, _ = json.Marshal("error create user")

	}
	if userInfo.UserEmail == "" || userInfo.Password == "" || emptyUser != 0 {
		log.Println("error: handlers/authandler CreateUserH() method SelectUSerIdMail() ")
		message, _ = json.Marshal(&MessageError{Message: "invalid params"})

		return
	}
	// если нет пользователя добавлем его в базу и отправлем письмо, условныи оператор - костыль, надо переписать

	hashPassword, _ := HashPassword(userInfo.Password)

	generateActivUuid, err := rep.Psql.CreateUser(r.Context(), userInfo.UserEmail, hashPassword, false)
	if err != nil {
		log.Println("error: handlers/authandler CreateUserH() method CreateUser")
		message, _ = json.Marshal(&MessageError{Message: "user is empty"})

		return
	}

	// путь убрать в .env
	strUrl := confirmURL + generateActivUuid
	statusEmail, err = SendEmailToConfirm(userInfo.UserEmail, strUrl)
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "smtp server error"})
		log.Println("letter not sent status email: ", statusEmail)

		return
	}
	log.Println("message sent")
	message, _ = json.Marshal(&MessageError{Message: "check you email"})
}

// переход по ссылке из письма
func (rep *AuthInquirysRepository) EmailActivateH(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		message []byte
	)

	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	index := vars["uuid"]

	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(message)
			log.Println("message not send")
		} else {
			log.Println("send message")
			w.Write(message)
		}
	}()

	err = rep.Psql.CheckUserUuidToEmail(r.Context(), index)
	if err != nil {
		log.Println("error: handlers/authandler EmailActivateH() method CheckUserUuidToEmail()")
		message, _ = json.Marshal(&MessageError{Message: "error message"})

		return
	}
	message, _ = json.Marshal(&MessageError{Message: "user is active"})
}

// 1) Проверить наличие пользователя в базе, правильность введеных логина и пороля
// 2) Если 1 этап успешен проверить поле activate если true - выдать токен, если не активен отправить письмо на почту, удалив предыдущу запись в emailactivate
// 3) миделваром обернуть остальные эндпоинты
// {
// "usermail":"you mail",
// "password":"pass"
// }
func (rep *AuthInquirysRepository) AuthentificateUserH(w http.ResponseWriter, r *http.Request) {
	var (
		userInfo UserInfoStruct
		message  []byte
		err      error
	)

	defer func() {
		if err != nil {
			log.Println(err, "Error request")
			w.WriteHeader(400)
			w.Write(message)
		} else {
			w.Write(message)
		}
	}()

	w.Header().Set("Content-Type", "application/json")

	err = json.NewDecoder(r.Body).Decode(&userInfo)
	if err != nil {
		log.Println("error decode body handler/authhandler AuthentificateUserH() decode body")
		message, _ = json.Marshal(&MessageError{Message: "error: body is empty"})

		return
	}
	log.Println(userInfo)

	user, err := rep.Psql.SelectUserByUserEmail(r.Context(), userInfo.UserEmail)
	if err != nil {
		log.Println("user is empty")
		message, _ = json.Marshal(&MessageError{Message: "user does not exist"})

		return
	}

	if user.UserEmail == "" || r.Body == http.NoBody {
		// если пользователя нет, то статус 400
		w.WriteHeader(http.StatusBadRequest)
		// TODO -  добавить обработку ошибки
		return
	}
	// проверка пароля
	checkPass := bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(userInfo.Password))
	if checkPass != nil {
		w.WriteHeader(http.StatusBadRequest)
		// TODO -  добавить обработку ошибки
		return
	}

	if !user.Activate {
		log.Println("user не активировал аккаунт")
		message, _ = json.Marshal(&MessageError{Message: "check mail and activate account"})

		return
	}
	// FIXME add error handler
	accessToken, _ := GenerateAcceessToken(user.UserId, user.UserEmail, user.Activate)
	refreshToken, _ := GenerateRefreshToken(user.UserId, user.Activate)

	// добавить токен в редис
	// Собрать в json и отправить
	// access и refresh tokens записывать в разные таблички

	// access token
	err = rep.Redis.AddAccessToken(r.Context(), user.UserEmail, accessToken)
	if err != nil {
		log.Println("error AddAccessToken")
		// TODO -  добавить обработку ошибки
		return
	}
	// refresh token
	err = rep.Redis.AddRefreshToken(r.Context(), user.UserEmail, refreshToken)
	if err != nil {
		log.Println("error AddRefreshToken")
		// TODO -  добавить обработку ошибки
		return
	}

	// return pair access and refresh tokens
	message, err = json.Marshal(AccessRefreshToken(accessToken, refreshToken))
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "error: marshal "})
		log.Println("error MArshal AcessRefreshToken()")

		return
	}
}

// FIXME все типы в начале файла -> затем экспортируемые функции -> затем не экспортируемые функции
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
		message []byte
		err     error
	)
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err != nil {
			log.Println(err, "Error xtyyuyg")
			w.WriteHeader(400)
			w.Write(message)
		} else {
			w.Write(message)
		}
		// FIXME зачем тут конструкция if else есл можно так
		// if err != nil {
		// 	log.Println(err, "Error xtyyuyg")
		// 	w.WriteHeader(400)
		// }
		// w.Write(message)

	}()

	var rT RefreshTokenStruct // FIXME отступ
	err = json.NewDecoder(r.Body).Decode(&rT)
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "body is empty"})
		log.Println("error: handlers/authandler RefreshTokenH() Decode()")

		return
	}
	// TODO
	token, err := jwt.Parse(rT.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrap(err, "token not parsed")
		}

		return MySignedAccessRefreshToken, err
	})
	if err != nil {
		log.Println("error auth")
		message, _ = json.Marshal(&MessageError{Message: "unauth"})
		return
	}

	log.Println(token)

	userId := int(token.Claims.(jwt.MapClaims)["userId"].(float64))
	tokenExp := int(token.Claims.(jwt.MapClaims)["exp"].(float64))
	timeNow := int(time.Now().Unix())

	// проверка наличия пользователя по id
	checkUserId, err := rep.Psql.SelectByUserId(r.Context(), userId)
	if err != nil {
		log.Println("error handler/authandler RefreshTokenH() in SelectByUserId()")
		message, _ = json.Marshal(&MessageError{Message: "no empty user or not valid token"})

		return
	}

	log.Println(checkUserId)

	checkRefreshToken, err := rep.Redis.GetRefreshTokenByUserEmail(r.Context(), checkUserId.UserEmail)
	if err != nil {
		log.Println("error handler/authandler RefreshTokenH() in GetRefreshTokenByUSerEmail()")
		message, _ = json.Marshal(&MessageError{Message: "no empty user or not valid token"})

		return
	}
	log.Println(checkRefreshToken)

	// // // если присланныи refresh токен валидныи и не протух, генерим новую пару, иначе  сообщаем, что токен не валиден
	if checkUserId.UserId == 0 || tokenExp < timeNow || rT.RefreshToken != checkRefreshToken {
		w.WriteHeader(http.StatusBadRequest)
		message, _ = json.Marshal(&MessageError{Message: "no empty user or not valid token"})

		return
	}

	// // генерация токенов
	accessToken, _ := GenerateAcceessToken(checkUserId.UserId, checkUserId.UserEmail, checkUserId.Activate)
	refreshToken, _ := GenerateRefreshToken(checkUserId.UserId, checkUserId.Activate)
	// // добавление access токена в redis

	err = rep.Redis.AddAccessToken(r.Context(), checkUserId.UserEmail, accessToken)
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "error add access token"})
		log.Println("error AddAccessToken() in handler/authhandler RefreshTokenH() ")

		return
	}

	err = rep.Redis.AddRefreshToken(r.Context(), checkUserId.UserEmail, refreshToken)
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "error add refresh token"})
		log.Println("error AddRefreshToken() in handler/authhandler RefreshTokenH() ")

		return
	}

	message, _ = json.Marshal(&AccessAndRefreshToken{AccessToken: accessToken, RefreshToken: refreshToken})
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "error response tokens"})
		log.Println("error marshal")

		return
	}

}

// запрос на смену парол
// пользователь вводит свои пароль
type CheckUserStruct struct {
	Email string `json:"useremail"`
}

func (rep *AuthInquirysRepository) ForgotPasswordH(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // FIXME Почемы ты сетишь заголовок ПЕРЕД объявлением переменны?
	var (
		message        []byte
		err            error
		checkUser      CheckUserStruct
		checkUserEmail *repository.UserInfoResponseStruct
		resetToken     string
	) // FIXME отступ
	defer func() {
		if err != nil {
			log.Println("error handler/authhandler ForgotPasswordH()")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(message)
		} else {
			w.Write(message)
		}
	}()

	// если пользователь есть отправить письмо на почту, если нет сообщить что пользователя нет
	err = json.NewDecoder(r.Body).Decode(&checkUser)
	if err != nil {
		log.Println("error decode data")
		message, _ = json.Marshal(&MessageError{Message: "invalid data handler/authhandler ForgotPasswordH()"})

		return
	}

	checkUserEmail, err = rep.Psql.SelectUserByUserEmail(r.Context(), checkUser.Email)
	if err != nil {
		log.Println("error SelectUserByUserEmail() in handler/authhandler ForgotPasswordH()")
		message, _ = json.Marshal(&MessageError{Message: "user does not exist"})

		return
	}

	resetToken, err = GenerateResetToken(checkUserEmail.UserEmail, checkUserEmail.UserId)
	if err != nil {
		log.Println("error: handlers/authhandler ForgotPasswordH() method GenerateResetToken()")
		message, _ = json.Marshal(&MessageError{Message: "reset token generation error"})

		return
	}

	if resetToken == "" {
		log.Println("error generate token")
		message, _ = json.Marshal(&MessageError{Message: "token not generate"})

		return
	}

	log.Println("generated reset token : ", resetToken)

	err = rep.Redis.AddResetToken(r.Context(), resetToken, checkUserEmail.UserEmail)
	if err != nil {
		log.Println("error in handlers/authhandler/ForgotPasswordH() AddRedisToken() ")
		message, _ = json.Marshal(&MessageError{Message: "error validation token"})

		return
	}
	// отправить письмо на почту
	err = SendEmailToPassReset(checkUserEmail.UserEmail, resetToken)
	if err != nil {
		log.Println("error  in handlers/authhandler/ForgotPasswordH() SendEmailToPassReset() method")
		message, _ = json.Marshal(&MessageError{Message: "error send message"})

		return
	}

	message, err = json.Marshal(&MessageError{Message: "a letter has been sent to your mail"})
	if err != nil {
		log.Println("error  in marshal json handler/authhandler SendEmailToPassReset() method")
		message, _ = json.Marshal(&MessageError{Message: "error send message"})

		return
	}
}

type ChangePassStruct struct {
	OriginPass  string `json:"origin_pass"`
	ConfirmPass string `json:"confirm_pass"`
}

func (rep *AuthInquirysRepository) ResetPassH(w http.ResponseWriter, r *http.Request) {
	// проверяем url и токен из письма
	// сверили токен с ключом из redis
	// если успешно, то получили значение по ключу
	// проверили переданные original pass и confirm pass
	// если они совпали, то хешируем original pass
	// обновление пароля по пользователю
	//
	w.Header().Set("contet-type", "applicateion/json")
	var (
		changePass ChangePassStruct
		message    []byte
		err        error
	)
	defer func() {
		// FIXME if else не нужен
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(message)
		} else {
			w.Write(message)
		}
	}()

	vars := mux.Vars(r)
	resetToken := vars["resetToken"]

	userEmailFromRedis, err := rep.Redis.GetResetTokenForCheckUserEmail(r.Context(), resetToken)
	if err != nil {
		log.Println("invalid token GetResetTokenForCheckUserEmail() in hendlers/authhandler ResetPassH()")
		message, _ = json.Marshal(&MessageError{Message: "invalid request"})

		return
	}

	err = json.NewDecoder(r.Body).Decode(&changePass)
	if err != nil {
		log.Println("error body")
		message, _ = json.Marshal(&MessageError{Message: "invalid data"})

		return
	}

	if changePass.ConfirmPass != changePass.OriginPass {
		log.Println("passwords must be different")
		message, _ = json.Marshal(&MessageError{Message: "passwords must be different"})

		return
	}

	hashOriginPass, err := HashPassword(changePass.OriginPass)
	if err != nil {
		log.Println("password hash error")
		message, _ = json.Marshal(&MessageError{Message: "error update pass"})

		return
	}

	err = rep.Psql.ChangePass(r.Context(), userEmailFromRedis, hashOriginPass)
	if err != nil {
		log.Println("error update ChangePass() in to handlers/authhandler")
		message, _ = json.Marshal(&MessageError{Message: "an error occurred while updating the password"})

		return
	}

	// FIXME везде ошибки Marshal обработай пример: ("encode error: %w", err)
	message, _ = json.Marshal(&MessageError{Message: "successfuljt password reset "})

}

// func (rep *AuthInquirysRepository) LogoutH(w http.ResponseWriter, r *http.Request) {
// 	// полученныи токен поместить в black list до истечения

// }
