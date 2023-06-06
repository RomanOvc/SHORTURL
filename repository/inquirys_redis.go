package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type RedisClient struct {
	redisDbTable0, redisDbTable1, redisDbTable2 *redis.Client
}

func NewRedisReposiory(redisDBTable0, redisDBTable1, redisDbTable2 *redis.Client) *RedisClient {
	return &RedisClient{redisDbTable0: redisDBTable0, redisDbTable1: redisDBTable1, redisDbTable2: redisDbTable2}
}

var (
	// FIXME добавить маски для записей в редисе и оставть ОДИН коннект (бд(таблицу))
	tokensMask = "tokensMask:%s"
)

type RedisInqurysInterface interface {
	// RedisTable 0
	AddAccessToken(ctx context.Context, userEmail, accessToken string) error
	GetAccessTokenByUserEmail(ctx context.Context, userEmail string) (string, error)

	// RedisTable 1
	AddRefreshToken(ctx context.Context, userEmail, refreshToken string) error
	GetRefreshTokenByUserEmsil(ctx context.Context, userEmail string) (string, error)

	// RedisTable 2
	AddResetToken(ctx context.Context, userEmail, resetToken string) error
	// GetResetTokenBy
}

// вставить access token  и username
// ["access_token"] = usermail
func (r *RedisClient) AddAccessToken(ctx context.Context, userEmail, accessToken string) error {
	err := r.redisDbTable0.Set(ctx, fmt.Sprintf(tokensMask, userEmail), accessToken, time.Minute*15).Err()
	if err != nil {
		errors.Wrapf(err, "error 'set comand to redis' repository/inqurys_redis  AddAccessAndRefreshToken()")
	}
	// FIXME "ok"
	return err
}

// return access_token
func (r *RedisClient) GetAccessTokenByUsermail(ctx context.Context, usermail string) (string, error) {
	var acceessToken string
	err := r.redisDbTable0.Get(ctx, usermail).Scan(&acceessToken)
	if err != nil {
		errors.Wrapf(err, "error 'get comand to redis' repository/inqurys_redis  GetAccessTokenByUsermail()")
	}
	return acceessToken, err
}

// add ["user_email"] = refres_htoken
func (r *RedisClient) AddRefreshToken(ctx context.Context, userEmail, refreshToken string) error {
	err := r.redisDbTable1.Set(ctx, userEmail, refreshToken, time.Hour*48).Err()
	if err != nil {
		errors.Wrap(err, "error 'set comand to redis' repository/inqurys_redis AddRefreshToken()")
	}

	return err
}

// return refresh_token
func (r *RedisClient) GetRefreshTokenByUserEmail(ctx context.Context, userEmail string) (string, error) {
	var refreshToken string
	err := r.redisDbTable1.Get(ctx, userEmail).Scan(&refreshToken)
	if err != nil {
		errors.Wrapf(err, "error 'get comand to redis' repository/inqurys_redis GetRefreshTokenByUSerEmail()")
	}

	return refreshToken, err
}

// add ["reset_token"] = user_Email
func (r *RedisClient) AddResetToken(ctx context.Context, resetToken, userEmail string) error {
	err := r.redisDbTable2.Set(ctx, resetToken, userEmail, time.Minute*5).Err()
	if err != nil {
		errors.Wrap(err, "error 'set command to redis' repository/inqurys_redis AddresetToken()")
	}

	return err
}

// return Reset_token
func (r *RedisClient) GetResetTokenForCheckUserEmail(ctx context.Context, resetToken string) (string, error) {
	var userEmail string // FIXME отступ
	err := r.redisDbTable2.Get(ctx, resetToken).Scan(&userEmail)
	if err != nil {
		errors.Wrapf(err, "error 'get comand to redis' repository/inqurys_redis GetRefreshTokenByUSerEmail()")
	}

	return userEmail, err
}
