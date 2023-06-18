package repository

import (
	"appurl/models"
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

func (r *AuthInquirysRepository) CreateUser(ctx context.Context, useremail, password string, activate bool) (string, error) {
	var userId int

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", errors.Wrap(err, "err begin transaction")
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (useremail,password,activate, created_user)values ($1, $2, $3, $4) RETURNING user_id`,
		useremail, password, activate, time.Now()).Scan(&userId)
	if err != nil {
		return "", errors.Wrap(err, "err insert into users")
	}

	userUuid := uuid.NewString()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO emailactivate (uid,user_id,active_until) values ($1, $2,$3)`,
		userUuid, userId, time.Now().Add(time.Hour))
	if err != nil {
		return "", errors.Wrap(err, "err insert into emailactivate")
	}

	tx.Commit()

	return userUuid, nil
}

func (r *AuthInquirysRepository) UserActivation(ctx context.Context, uuid string) error {
	var resUserId int

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "Begin transaction")
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	timeNow := time.Now()

	err = tx.QueryRowContext(ctx,
		`SELECT 
		    user_id 
		FROM 
		    emailactivate 
		WHERE 
		    uid = $1 and active_until > $2`,
		uuid, timeNow).Scan(&resUserId)
	if err != nil {
		return errors.Wrap(err, " CheckUserUuidToEmail()")
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE 
		    users 
		SET 
		    activate = $2 
		WHERE 
		    user_id = $1`,
		resUserId, true)
	if err != nil {
		return errors.Wrap(err, "UpdateActivateStatus()")
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM emailactivate WHERE uid =$1`, uuid)
	if err != nil {
		return errors.Wrap(err, "UpdateActivateStatus()")
	}

	tx.Commit()

	return nil

}

func (r *AuthInquirysRepository) SelectUserIdByMail(ctx context.Context, userEmail string) (int, error) {
	var userId int

	ro := r.db.QueryRowContext(ctx, `SELECT user_id from users where useremail = $1;`, userEmail)
	switch err := ro.Scan(&userId); err {
	case nil:
		return userId, nil
	case sql.ErrNoRows:
		return 0, err
	default:
		return 0, err
	}
}

func (r *AuthInquirysRepository) SelectUserByUserEmail(ctx context.Context, userEmail string) (*models.UserInfoResponseStruct, error) {
	var userInfoRep models.UserInfoResponseStruct

	ro := r.db.QueryRowContext(ctx, `SELECT user_id, useremail, password,activate FROM users where useremail = $1`, userEmail)
	err := ro.Scan(&userInfoRep.UserId, &userInfoRep.UserEmail, &userInfoRep.Pass, &userInfoRep.Activate)
	if err != nil {
		return nil, errors.Wrap(err, "UpdateActivateStatus()")
	}

	return &userInfoRep, nil
}

func (r *AuthInquirysRepository) SelectByUserId(ctx context.Context, userId int) (*models.UserInfoResponseStruct, error) {
	var userInfo models.UserInfoResponseStruct
	row := r.db.QueryRowContext(ctx, `SELECT user_id, useremail from users where user_id = $1`, userId)
	err := row.Scan(&userInfo.UserId, &userInfo.UserEmail)
	if err != nil {
		return nil, errors.Wrap(err, "SelectByUserId()")
	}

	return &userInfo, nil
}

// hash origin pass
func (r *AuthInquirysRepository) ChangePass(ctx context.Context, userEmail, hashOriginPass string) error {
	err := r.db.QueryRowContext(ctx, `UPDATE users SET password=$1 where usermail=$2`, hashOriginPass, userEmail).Err()
	if err != nil {
		return errors.Wrap(err, "update error")
	}

	return nil
}

func (r *AuthInquirysRepository) CheckEmailActivate(ctx context.Context, user_id int) (string, error) {
	var uid string

	ro := r.db.QueryRowContext(ctx, `SELECT uid FROM emailactivate where user_id = $1 and active_until > $2`, user_id, time.Now().Format("2006-01-02 15:04:05"))
	switch err := ro.Scan(&uid); err {
	case nil:
		return uid, nil
	case sql.ErrNoRows:
		return "", err
	default:
		return "", err
	}
}

func (r *AuthInquirysRepository) InsertUidForEmailActivate(ctx context.Context, userId int) (string, error) {
	var uid string

	ro := r.db.QueryRowContext(ctx, `INSERT INTO emailactivate (uid,user_id,active_until) values ($1, $2,$3) RETURNING uid`, uuid.NewString(), userId, time.Now().Add(time.Hour*24))
	err := ro.Scan(&uid)
	if err != nil {
		return "", errors.Wrap(err, "repository/authinquirys UpdateActivateStatus() method error")
	}
	return uid, nil

}
