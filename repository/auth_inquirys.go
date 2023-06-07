package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type AuthInquirysRepository struct {
	db *sql.DB
}

func NewAuthInquirysRepository(db *sql.DB) *AuthInquirysRepository {
	return &AuthInquirysRepository{db: db}
}

type UserInfoResponseStruct struct {
	UserId       int
	UserEmail    string
	Pass         string
	Activate     bool
	RefreshToken string
}

func (r *AuthInquirysRepository) CreateUser(ctx context.Context, usermail, password string, activate bool) (string, error) {
	var userId int

	tx, err := r.db.BeginTx(ctx, nil)

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	createTime := time.Now()
	err = tx.QueryRowContext(ctx, "INSERT INTO users (usermail, password, activate, create_time_user) VALUES ($1, $2, $3, $4) RETURNING user_id", usermail, password, activate, createTime).Scan(&userId)
	if err != nil {
		return "", errors.Wrap(err, "err insert into users")
	}

	userUuid := uuid.NewString()

	_, err = r.db.ExecContext(ctx, "INSERT INTO emailactivate (uid, user_id, active_until) VALUES ($1, $2,$3)", userUuid, userId, time.Now().Add(time.Hour))
	if err != nil {
		return "", errors.Wrap(err, "err insert into emailactivate")
	}

	tx.Commit()

	return userUuid, err
}

func (r *AuthInquirysRepository) CheckUserUuidToEmail(ctx context.Context, uuid string) error {

	var resUserId int

	tx, err := r.db.BeginTx(ctx, nil)

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	timeNow := time.Now()

	err = tx.QueryRowContext(ctx, "select user_id from emailactivate where uid = $1 and active_until > $2", uuid, timeNow).Scan(&resUserId)
	if err != nil {
		return errors.Wrap(err, "repository/authinquirys  CheckUserUuidToEmail() method error")
	}

	_, err = r.db.ExecContext(ctx, "UPDATE users SET activate = $2 WHERE user_id = $1", resUserId, true)
	if err != nil {
		return errors.Wrap(err, "repository/authinquirys UpdateActivateStatus() method error")
	}
	tx.Commit()
	return err

}

func (r *AuthInquirysRepository) UpdateActivateStatus(ctx context.Context, usermail string) (int, error) {
	// добавит удаление записи из табли emailactivate
	var userId int
	tx, err := r.db.BeginTx(ctx, nil)

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = r.db.QueryRowContext(ctx, "UPDATE users SET activate = $2 WHERE usermail = $1 RETURNING user_id", usermail, true).Scan(&userId)
	if err != nil {
		errors.Wrap(err, "repository/authinquirys UpdateActivateStatus() method error")
	}

	_, err = r.db.ExecContext(ctx, "DELETE FROM emailactivate WHERE user_id =$1", userId)
	if err != nil {
		errors.Wrap(err, "repository/authinquirys UpdateActivateStatus() method error")
	}
	tx.Commit()
	return userId, err

}

// ????????????????????

func (r *AuthInquirysRepository) SelectUserIdByMail(ctx context.Context, userEmail string) (int, error) {
	var userId int

	err := r.db.QueryRowContext(ctx, "SELECT user_id from users where usermail = $1 ", userEmail).Scan(&userId)
	// FIXME handle errorЫ
	if err != nil {
		return 0, errors.Wrap(err, "select user id by email")
	}

	return userId, err
}

// позже будет возвращть объект пользователя, для создания токена
func (r *AuthInquirysRepository) SelectUserByUserEmail(ctx context.Context, userEmail string) (*UserInfoResponseStruct, error) {
	var userInfoRep UserInfoResponseStruct
	ro := r.db.QueryRowContext(ctx, "SELECT user_id, usermail, password,activate FROM users where usermail = $1", userEmail)
	err := ro.Scan(&userInfoRep.UserId, &userInfoRep.UserEmail, &userInfoRep.Pass, &userInfoRep.Activate)
	if err != nil {
		errors.Wrap(err, "repository/auth_inquirys UpdateActivateStatus() method error")
	}
	return &userInfoRep, err
}

func (r *AuthInquirysRepository) SelectByUserId(ctx context.Context, userId int) (*UserInfoResponseStruct, error) {
	// select в табл. по user_id.
	// return { user_id,usermail}
	var userInfo UserInfoResponseStruct
	row := r.db.QueryRowContext(ctx, "SELECT user_id, usermail from users where user_id = $1", userId)
	err := row.Scan(&userInfo.UserId, &userInfo.UserEmail)
	if err != nil {
		errors.Wrap(err, "repository/auth_inquirys SelectByUserId() method error")
	}
	return &userInfo, err
}

// hash origin pass
func (r *AuthInquirysRepository) ChangePass(ctx context.Context, userEmail, hashOriginPass string) error {
	err := r.db.QueryRowContext(ctx, "UPDATE users SET password=$1 where usermail=$2", hashOriginPass, userEmail).Err()
	if err != nil {
		errors.Wrap(err, "asdasdasdasdadasdasdasd")
	}
	return err

}
