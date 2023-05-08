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
	Usermail     string
	Pass         string
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

	err = tx.QueryRowContext(ctx, "INSERT INTO users (usermail, password, activate) VALUES ($1, $2,$3) RETURNING user_id", usermail, password, activate).Scan(&userId)
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
// FIXME maybe select user id by email
func (r *AuthInquirysRepository) SelectUserMail(ctx context.Context, userEmail string) int {
	var userId int

	_ = r.db.QueryRowContext(ctx, "SELECT user_id from users where usermail = $1 ", userEmail).Scan(&userId)
	// FIXME handle error
	// if err != nil {
	// 	return errors.Wrap(err, "select user id by email")
	// }

	return userId
}

// позже будет возвращть объект пользователя, для создания токена
func (r *AuthInquirysRepository) SelectUser(ctx context.Context, userEmail string) *UserInfoResponseStruct {
	var userInfoRep UserInfoResponseStruct
	row := r.db.QueryRowContext(ctx, "SELECT user_id, usermail, password FROM users where usermail = $1", userEmail)
	row.Scan(&userInfoRep.UserId, &userInfoRep.Usermail, &userInfoRep.Pass)
	return &userInfoRep
}

func (r *AuthInquirysRepository) SelectByUserId(ctx context.Context, userId int) *UserInfoResponseStruct {
	// select в табл. по user_id.
	var userInfo UserInfoResponseStruct
	row := r.db.QueryRowContext(ctx, "SELECT user_id, usermail, refresh_token from users where user_id = $1", userId)
	row.Scan(&userInfo.UserId, &userInfo.Usermail, &userInfo.RefreshToken)
	return &userInfo
}

func (r *AuthInquirysRepository) UpdateRefershTokenForUser(ctx context.Context, userId int, refreshToken string) (int, error) {
	var userIdDb int
	err := r.db.QueryRowContext(ctx, "UPDATE users SET refresh_token=$1 where user_id = $2 RETURNING user_id ", refreshToken, userId).Scan(&userIdDb)
	if err != nil {
		return userIdDb, errors.Wrap(err, "repository/authinquirys UpdateRefershTokenForUser() method error")
	}
	return userIdDb, nil

}
