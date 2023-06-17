package handlers

import (
	"appurl/models"
	"appurl/repository"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

const (
	confirmURL = "http://127.0.0.1:8001/auth/confirm/"
)

type authInquirysRepository struct {
	Psql  *repository.AuthInquirysRepository
	Redis *repository.RedisClient
}

func NewAuthInquirysRepository(postgres *repository.AuthInquirysRepository, redis *repository.RedisClient) *authInquirysRepository {
	return &authInquirysRepository{Psql: postgres, Redis: redis}
}

func (rep *authInquirysRepository) CreateUserH(w http.ResponseWriter, r *http.Request) {
	var (
		userInfo    models.UserInfoStruct
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
		log.Printf("Decode: %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "invalid params"})

		return
	}

	emptyUser, err := rep.Psql.SelectUserIdByMail(r.Context(), userInfo.UserEmail)
	if err != nil {
		log.Printf("SelectUserIdByMail(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "невозможно добавить"})

		return
	}

	if userInfo.UserEmail == "" || userInfo.Password == "" || emptyUser != 0 {
		err = fmt.Errorf("body is empty")
		message, _ = json.Marshal(&models.MessageError{Message: "user is empty  or invalid pararmetrs"})

		return
	}

	hashPassword, err := HashPassword(userInfo.Password)
	if err != nil {
		err = fmt.Errorf("password is empty")
		return
	}

	generateActivUuid, err := rep.Psql.CreateUser(r.Context(), userInfo.UserEmail, hashPassword, false)
	if err != nil {
		log.Printf("CreateUser(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "user is empty"})

		return
	}

	statusEmail, err = SendEmailToConfirm(userInfo.UserEmail, confirmURL+generateActivUuid)
	if err != nil {
		message, _ = json.Marshal(&models.MessageError{Message: "smtp server error"})
		log.Printf("SendEmailToConfirm: %s, status send email: %t", err.Error(), statusEmail)

		return
	}

	message, _ = json.Marshal(&models.MessageError{Message: "check you email"})

}

// переход по ссылке из письма
func (rep *authInquirysRepository) EmailActivateH(w http.ResponseWriter, r *http.Request) {
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
			log.Println("message not send")
		}
		w.Write(message)
	}()

	err = rep.Psql.CheckUserUuidToEmail(r.Context(), index)
	if err != nil {
		log.Printf("CheckUserUuidToEmail(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error message"})

		return
	}

	message, _ = json.Marshal(&models.MessageError{Message: "user is active"})
}

func (rep *authInquirysRepository) AuthentificateUserH(w http.ResponseWriter, r *http.Request) {
	var (
		userInfo models.UserInfoStruct
		message  []byte
		err      error
	)

	defer func() {
		if err != nil {
			log.Println(err, "Error request")
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(message)
	}()

	w.Header().Set("Content-Type", "application/json")

	err = json.NewDecoder(r.Body).Decode(&userInfo)
	if err != nil {
		log.Printf("Decode(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error: body is empty"})

		return
	}

	user, err := rep.Psql.SelectUserByUserEmail(r.Context(), userInfo.UserEmail)
	if err != nil {
		log.Printf("SelectUserByUserEmail(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "user does not exist"})

		return
	}

	if user.UserEmail == "" || r.Body == http.NoBody {
		err = fmt.Errorf("body is empty")
		return
	}

	// проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(userInfo.Password))
	if err != nil {
		log.Printf("CompareHashAndPassword(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "Error pass"})

		return
	}

	if !user.Activate {
		err = fmt.Errorf("user not active account")
		message, _ = json.Marshal(&models.MessageError{Message: "check mail and activate account"})

		return
	}

	accessToken, err := GenerateAcceessToken(user.UserId, user.UserEmail, user.Activate)
	if err != nil {
		log.Printf("GenerateAcceessToken(): %s", err.Error())

		return
	}

	refreshToken, err := GenerateRefreshToken(user.UserId, user.Activate)
	if err != nil {
		log.Printf("GenerateRefreshToken(): %s", err.Error())

		return
	}

	err = rep.Redis.AddAccessToken(r.Context(), user.UserEmail, accessToken)
	if err != nil {
		log.Printf("AddAccessToken(): %s", err.Error())

		return
	}
	// refresh token
	err = rep.Redis.AddRefreshToken(r.Context(), user.UserEmail, refreshToken)
	if err != nil {
		log.Printf("AddRefreshToken(): %s", err.Error())

		return
	}

	// return pair access and refresh tokens
	message, err = json.Marshal(AccessRefreshToken(accessToken, refreshToken))
	if err != nil {
		message, _ = json.Marshal(&models.MessageError{Message: "error: marshal "})
		log.Printf("Marshal(): %s", err.Error())

		return
	}
}

func (rep *authInquirysRepository) RefreshTokenH(w http.ResponseWriter, r *http.Request) {
	var (
		message []byte
		err     error
		rT      models.RefreshTokenStruct
	)

	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err != nil {
			log.Println(err, "RefreshTokenH()")
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(message)

	}()

	err = json.NewDecoder(r.Body).Decode(&rT)
	if err != nil {
		log.Printf("Decode(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "body is empty"})

		return
	}

	token, err := TokenParse(rT.RefreshToken)
	if err != nil {
		log.Printf("TokenParse(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "unauth"})

		return
	}

	userId := int(token.Claims.(jwt.MapClaims)["userId"].(float64))
	tokenExp := int(token.Claims.(jwt.MapClaims)["exp"].(float64))
	timeNow := int(time.Now().Unix())

	// проверка наличия пользователя по id
	checkUserId, err := rep.Psql.SelectByUserId(r.Context(), userId)
	if err != nil {
		log.Printf("SelectByUserId(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "no empty user or not valid token"})

		return
	}

	checkRefreshToken, err := rep.Redis.GetRefreshTokenByUserEmail(r.Context(), checkUserId.UserEmail)
	if err != nil {
		log.Printf("GetRefreshTokenByUserEmail(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "no empty user or not valid token"})

		return
	}

	// // // если присланныи refresh токен валидныи и не протух, генерим новую пару, иначе  сообщаем, что токен не валиден
	if checkUserId.UserId == 0 || tokenExp < timeNow || rT.RefreshToken != checkRefreshToken {
		err = fmt.Errorf("token not valid")
		message, _ = json.Marshal(&models.MessageError{Message: "no empty user or not valid token"})

		return
	}

	// // генерация токенов
	accessToken, err := GenerateAcceessToken(checkUserId.UserId, checkUserId.UserEmail, checkUserId.Activate)
	if err != nil {
		log.Printf("GenerateAcceessToken(): %s", err.Error())

		return
	}

	refreshToken, err := GenerateRefreshToken(checkUserId.UserId, checkUserId.Activate)
	if err != nil {
		log.Printf("GenerateRefreshToken(): %s", err.Error())

		return
	}

	err = rep.Redis.AddAccessToken(r.Context(), checkUserId.UserEmail, accessToken)
	if err != nil {
		log.Printf("AddAccessToken(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error add access token"})

		return
	}

	err = rep.Redis.AddRefreshToken(r.Context(), checkUserId.UserEmail, refreshToken)
	if err != nil {
		log.Printf("AddRefreshToken(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error add refresh token"})

		return
	}

	message, _ = json.Marshal(&AccessAndRefreshToken{AccessToken: accessToken, RefreshToken: refreshToken})
	if err != nil {
		log.Printf("Marshal(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error response tokens"})

		return
	}
}

func (rep *authInquirysRepository) ForgotPasswordH(w http.ResponseWriter, r *http.Request) {
	var (
		message        []byte
		err            error
		checkUser      models.CheckUserStruct
		checkUserEmail *models.UserInfoResponseStruct
		resetToken     string
	)

	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err != nil {
			log.Println("handler/authhandler ForgotPasswordH()")
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(message)
	}()

	// если пользователь есть отправить письмо на почту, если нет сообщить что пользователя нет
	err = json.NewDecoder(r.Body).Decode(&checkUser)
	if err != nil {
		log.Printf("Decode(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "invalid data handler/authhandler ForgotPasswordH()"})

		return
	}

	checkUserEmail, err = rep.Psql.SelectUserByUserEmail(r.Context(), checkUser.Email)
	if err != nil {
		log.Printf("SelectUserByUserEmail(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "user does not exist"})

		return
	}

	resetToken, err = GenerateResetToken(checkUserEmail.UserEmail, checkUserEmail.UserId)
	if err != nil {
		log.Printf("GenerateResetToken(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "reset token generation error"})

		return
	}

	if resetToken == "" {
		err = fmt.Errorf("not valid reset token")
		message, _ = json.Marshal(&models.MessageError{Message: "token not generate"})

		return
	}

	err = rep.Redis.AddResetToken(r.Context(), resetToken, checkUserEmail.UserEmail)
	if err != nil {
		log.Printf("AddResetToken(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error validation token"})

		return
	}
	// отправить письмо на почту
	err = SendEmailToPassReset(checkUserEmail.UserEmail, resetToken)
	if err != nil {
		log.Printf("SendEmailToPassReset(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error send message"})

		return
	}

	message, err = json.Marshal(&models.MessageError{Message: "a letter has been sent to your mail"})
	if err != nil {
		log.Printf("Marshal(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error send message"})

		return
	}
}

func (rep *authInquirysRepository) ResetPassH(w http.ResponseWriter, r *http.Request) {
	var (
		changePass models.ChangePassStruct
		message    []byte
		err        error
	)

	w.Header().Set("contet-type", "applicateion/json")

	defer func() {
		if err != nil {
			log.Println("ResetPassH(): ")
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(message)
	}()

	vars := mux.Vars(r)
	resetToken := vars["resetToken"]

	userEmailFromRedis, err := rep.Redis.GetResetTokenForCheckUserEmail(r.Context(), resetToken)
	if err != nil {
		log.Printf("GetResetTokenForCheckUserEmail(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "invalid request"})

		return
	}

	err = json.NewDecoder(r.Body).Decode(&changePass)
	if err != nil {
		log.Printf("Decode(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "invalid data"})

		return
	}

	if changePass.ConfirmPass != changePass.OriginPass {
		err = fmt.Errorf("confirm pass not equal orign pass ")
		message, _ = json.Marshal(&models.MessageError{Message: "passwords must be different"})

		return
	}

	hashOriginPass, err := HashPassword(changePass.OriginPass)
	if err != nil {
		log.Printf("HashPassword(): %s", err)
		message, _ = json.Marshal(&models.MessageError{Message: "error update pass"})

		return
	}

	err = rep.Psql.ChangePass(r.Context(), userEmailFromRedis, hashOriginPass)
	if err != nil {
		log.Printf("ChangePass(): %s", err)
		message, _ = json.Marshal(&models.MessageError{Message: "an error occurred while updating the password"})

		return
	}

	message, _ = json.Marshal(&models.MessageError{Message: "successfull password reset "})
}
