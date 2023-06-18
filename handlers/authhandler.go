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
		userInfo       models.UserInfoStruct
		err            error
		statusEmail    bool
		messageForUser string
	)

	defer func() {
		var message []byte
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			message, _ = json.Marshal(&models.MessageResponse{Message: err.Error()})
		} else {
			message, _ = json.Marshal(&models.MessageResponse{Message: messageForUser})
		}
		w.Write(message)
	}()

	w.Header().Set("Content-Type", "application/json")

	err = json.NewDecoder(r.Body).Decode(&userInfo)
	if err != nil {
		log.Printf("Decode(): %s", err)

		return
	}

	if userInfo.UserEmail == "" || userInfo.Password == "" {
		err = fmt.Errorf("the request failed")

		return
	}

	hashPassword, err := HashPassword(userInfo.Password)
	if err != nil {
		log.Printf("HashPassword(): %s", err.Error())
		err = fmt.Errorf("password generation failed")

		return
	}

	generateActivUuid, err := rep.Psql.CreateUser(r.Context(), userInfo.UserEmail, hashPassword, false)
	if err != nil {
		log.Printf("CreateUser(): %s", err.Error())
		err = fmt.Errorf("user is empty")

		return
	}

	statusEmail, err = SendEmailToConfirm(userInfo.UserEmail, confirmURL+generateActivUuid)
	if err != nil {
		log.Printf("SendEmailToConfirm: %s, status send email: %t", err.Error(), statusEmail)
		err = fmt.Errorf("error send message")

		return
	}

	messageForUser = "check you email"
}

// переход по ссылке из письма
func (rep *authInquirysRepository) EmailActivateH(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		message        []byte
		messageForUser string
	)

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	uid := vars["uid"]

	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("EmailActivateH():")
			message, _ = json.Marshal(&models.MessageResponse{Message: err.Error()})
		} else {
			message, _ = json.Marshal(&models.MessageResponse{Message: messageForUser})
		}
		w.Write(message)
	}()

	err = rep.Psql.UserActivation(r.Context(), uid)
	if err != nil {
		log.Printf("UserActivation(): %s", err.Error())
		err = fmt.Errorf("the letter is outdated")

		return
	}

	messageForUser = "user is active"
}

func (rep *authInquirysRepository) AuthentificateUserH(w http.ResponseWriter, r *http.Request) {
	var (
		userInfo models.UserInfoStruct
		message  []byte
		err      error
	)

	defer func() {
		if err != nil {
			log.Println("AuthentificateUserH():")
			w.WriteHeader(http.StatusBadRequest)
			message, _ = json.Marshal(&models.MessageResponse{Message: err.Error()})
		}
		w.Write(message)
	}()

	w.Header().Set("Content-Type", "application/json")

	err = json.NewDecoder(r.Body).Decode(&userInfo)
	if err != nil {
		log.Printf("Decode(): %s", err.Error())

		return
	}

	if r.Body == http.NoBody || userInfo.UserEmail == "" || userInfo.Password == "" {
		err = fmt.Errorf("invalid parameter")

		return
	}

	user, err := rep.Psql.SelectUserByUserEmail(r.Context(), userInfo.UserEmail)
	if err != nil {
		log.Printf("SelectUserByUserEmail(): %s", err.Error())
		err = fmt.Errorf("incorrect username")

		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(userInfo.Password))
	if err != nil {
		log.Printf("CompareHashAndPassword(): %s", err.Error())
		err = fmt.Errorf("incorrect password")

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

	err = rep.Redis.AddRefreshToken(r.Context(), user.UserEmail, refreshToken)
	if err != nil {
		log.Printf("AddRefreshToken(): %s", err.Error())

		return
	}

	message, err = json.Marshal(AccessRefreshToken(accessToken, refreshToken))
	if err != nil {
		log.Printf("Marshal(): %s", err)
		err = fmt.Errorf("unknown error")

		return
	}
}

func (rep *authInquirysRepository) RepeatEmailActivateH(w http.ResponseWriter, r *http.Request) {
	var (
		message        []byte
		err            error
		useremail      models.CheckUserStruct
		messageForUser string
	)

	defer func() {
		if err != nil {
			log.Println("AuthentificateUserH():")
			w.WriteHeader(http.StatusBadRequest)
			message, _ = json.Marshal(&models.MessageResponse{Message: err.Error()})
		} else {
			message, _ = json.Marshal(&models.MessageResponse{Message: messageForUser})
		}
		w.Write(message)
	}()

	w.Header().Set("Content-Type", "application/json")

	err = json.NewDecoder(r.Body).Decode(&useremail)
	if err != nil {
		log.Printf("Decode(): %s", err)

		return
	}

	userId, err := rep.Psql.SelectUserIdByMail(r.Context(), useremail.Email)
	if err != nil {
		log.Printf("SelectUserIdMail(): %s", err.Error())
		err = fmt.Errorf("incorrect useremail")

		return
	}

	uuid, err := rep.Psql.CheckEmailActivate(r.Context(), userId)
	if err != nil {
		log.Printf("CheckEmailActivate(): %s", err.Error())
		err = fmt.Errorf("incorrect")

		return
	}

	if uuid == "" {
		uuid, err = rep.Psql.InsertUidForEmailActivate(r.Context(), userId)
		if err != nil {
			log.Printf("CheckEmailActivate(): %s", err.Error())
			err = fmt.Errorf("incorrect useremail")

			return
		}
	}

	statusEmail, err := SendEmailToConfirm(useremail.Email, confirmURL+uuid)
	if err != nil {
		log.Printf("SendEmailToConfirm: %s, status send email: %t", err.Error(), statusEmail)
		err = fmt.Errorf("error send message")

		return
	}

	messageForUser = "check you email"
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
			message, _ = json.Marshal(&models.MessageResponse{Message: err.Error()})
		}
		w.Write(message)

	}()

	err = json.NewDecoder(r.Body).Decode(&rT)
	if err != nil {
		log.Printf("Decode(): %s", err.Error())

		return
	}

	token, err := TokenParse(rT.RefreshToken)
	if err != nil {
		log.Printf("TokenParse(): %s", err.Error())
		err = fmt.Errorf("unauth")

		return
	}

	userId := int(token.Claims.(jwt.MapClaims)["userId"].(float64))
	tokenExp := int(token.Claims.(jwt.MapClaims)["exp"].(float64))
	timeNow := int(time.Now().Unix())

	checkUserId, err := rep.Psql.SelectByUserId(r.Context(), userId)
	if err != nil {
		log.Printf("SelectByUserId(): %s", err.Error())
		err = fmt.Errorf("no empty user or not valid token")

		return
	}

	checkRefreshToken, err := rep.Redis.GetRefreshTokenByUserEmail(r.Context(), checkUserId.UserEmail)
	if err != nil {
		log.Printf("GetRefreshTokenByUserEmail(): %s", err.Error())
		err = fmt.Errorf("no empty user or not valid token")

		return
	}

	if checkUserId.UserId == 0 || tokenExp < timeNow || rT.RefreshToken != checkRefreshToken {
		err = fmt.Errorf("token not valid")

		return
	}

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

		return
	}

	err = rep.Redis.AddRefreshToken(r.Context(), checkUserId.UserEmail, refreshToken)
	if err != nil {
		log.Printf("AddRefreshToken(): %s", err.Error())

		return
	}

	message, _ = json.Marshal(&AccessAndRefreshToken{AccessToken: accessToken, RefreshToken: refreshToken})
	if err != nil {
		log.Printf("Marshal(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageResponse{Message: "error response tokens"})

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
			log.Println("ForgotPasswordH()")
			w.WriteHeader(http.StatusBadRequest)
			message, _ = json.Marshal(&models.MessageResponse{Message: err.Error()})
		}
		w.Write(message)
	}()

	// если пользователь есть отправить письмо на почту, если нет сообщить что пользователя нет
	err = json.NewDecoder(r.Body).Decode(&checkUser)
	if err != nil {
		log.Printf("Decode(): %s", err.Error())

		return
	}

	checkUserEmail, err = rep.Psql.SelectUserByUserEmail(r.Context(), checkUser.Email)
	if err != nil {
		log.Printf("SelectUserByUserEmail(): %s", err.Error())
		err = fmt.Errorf("user does not exist")

		return
	}

	resetToken, err = GenerateResetToken(checkUserEmail.UserEmail, checkUserEmail.UserId)
	if err != nil {
		log.Printf("GenerateResetToken(): %s", err.Error())

		return
	}

	if resetToken == "" {
		err = fmt.Errorf("not valid reset token")

		return
	}

	err = rep.Redis.AddResetToken(r.Context(), resetToken, checkUserEmail.UserEmail)
	if err != nil {
		log.Printf("AddResetToken(): %s", err.Error())

		return
	}
	// отправить письмо на почту
	err = SendEmailToPassReset(checkUserEmail.UserEmail, resetToken)
	if err != nil {
		log.Printf("SendEmailToPassReset(): %s", err.Error())

		return
	}

	message, err = json.Marshal(&models.MessageResponse{Message: "a letter has been sent to your mail"})
	if err != nil {
		log.Printf("Marshal(): %s", err.Error())

		return
	}
}

func (rep *authInquirysRepository) ResetPassH(w http.ResponseWriter, r *http.Request) {
	var (
		changePass     models.ChangePassStruct
		message        []byte
		err            error
		messageForUser string
	)

	w.Header().Set("contet-type", "applicateion/json")

	defer func() {
		if err != nil {
			log.Println("ResetPassH(): ")
			w.WriteHeader(http.StatusBadRequest)
			message, _ = json.Marshal(&models.MessageResponse{Message: err.Error()})
		} else {
			message, _ = json.Marshal(&models.MessageResponse{Message: messageForUser})
		}
		w.Write(message)
	}()

	vars := mux.Vars(r)
	resetToken := vars["resetToken"]

	err = json.NewDecoder(r.Body).Decode(&changePass)
	if err != nil {
		log.Printf("Decode(): %s", err.Error())

		return
	}

	userEmailFromRedis, err := rep.Redis.GetResetTokenForCheckUserEmail(r.Context(), resetToken)
	if err != nil {
		log.Printf("GetResetTokenForCheckUserEmail(): %s", err.Error())

		return
	}

	if changePass.ConfirmPass != changePass.OriginPass {
		err = fmt.Errorf("confirm pass not equal orign pass ")

		return
	}

	hashOriginPass, err := HashPassword(changePass.OriginPass)
	if err != nil {
		log.Printf("HashPassword(): %s", err)

		return
	}

	err = rep.Psql.ChangePass(r.Context(), userEmailFromRedis, hashOriginPass)
	if err != nil {
		log.Printf("ChangePass(): %s", err)

		return
	}

	messageForUser = "successfull password reset "
}
