package repository

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type RedisClient struct {
	redisDbTable0 *redis.Client
}

// Заготовка для redis данных
var (
	prefixForAccessToken  = "acessToken:"
	prefixForRefreshToken = "refreshToken:"
	prefixForResetToken   = "resetToken:"
)

func NewRedisReposiory(redisDBTable0 *redis.Client) *RedisClient {
	return &RedisClient{redisDbTable0: redisDBTable0}
}

type RedisInqurysInterface interface {
	// Access Token
	AddAccessToken(ctx context.Context, userEmail, accessToken string) error
	GetAccessTokenByUserEmail(ctx context.Context, userEmail string) (string, error)

	// Refresh Token
	AddRefreshToken(ctx context.Context, userEmail, refreshToken string) error
	GetRefreshTokenByUserEmsil(ctx context.Context, userEmail string) (string, error)

	// Reset Token
	AddResetToken(ctx context.Context, userEmail, resetToken string) error
	GetResetTokenForCheckUserEmail(ctx context.Context, resetToken string) (string, error)
}

// вставить access token  и username
// ["access_token"] = usermail
func (r *RedisClient) AddAccessToken(ctx context.Context, userEmail, accessToken string) error {
	err := r.redisDbTable0.Set(ctx, prefixForAccessToken+userEmail, accessToken, time.Minute*15).Err()
	if err != nil {
		return errors.Wrapf(err, "error 'set comand to redis' repository/inqurys_redis  AddAccessAndRefreshToken()")
	}

	return nil
}

// return access_token
func (r *RedisClient) GetAccessTokenByUsermail(ctx context.Context, usermail string) (string, error) {
	var acceessToken string

	err := r.redisDbTable0.Get(ctx, prefixForAccessToken+usermail).Scan(&acceessToken)
	if err != nil {
		return "", errors.Wrapf(err, "error 'get comand to redis' repository/inqurys_redis  GetAccessTokenByUsermail()")
	}

	return acceessToken, err
}

// add ["user_email"] = refres_htoken
func (r *RedisClient) AddRefreshToken(ctx context.Context, userEmail, refreshToken string) error {
	err := r.redisDbTable0.Set(ctx, prefixForRefreshToken+userEmail, refreshToken, time.Hour*48).Err()
	if err != nil {
		return errors.Wrap(err, "error 'set comand to redis' repository/inqurys_redis AddRefreshToken()")
	}

	return nil
}

// return refresh_token
func (r *RedisClient) GetRefreshTokenByUserEmail(ctx context.Context, userEmail string) (string, error) {
	var refreshToken string

	err := r.redisDbTable0.Get(ctx, prefixForRefreshToken+userEmail).Scan(&refreshToken)
	if err != nil {
		return "", errors.Wrapf(err, "'get comand to redis' repository/inqurys_redis GetRefreshTokenByUSerEmail()")
	}

	return refreshToken, nil
}

func (r *RedisClient) AddResetToken(ctx context.Context, resetToken, userEmail string) error {
	modifiedResetToken := prefixForResetToken + resetToken

	err := r.redisDbTable0.Set(ctx, modifiedResetToken, userEmail, time.Minute*5).Err()
	if err != nil {
		return errors.Wrap(err, "error 'set command to redis' repository/inqurys_redis AddresetToken()")
	}

	return nil
}

func (r *RedisClient) GetResetTokenForCheckUserEmail(ctx context.Context, resetToken string) (string, error) {
	var userEmail string

	modifiedResetToken := prefixForResetToken + resetToken

	err := r.redisDbTable0.Get(ctx, modifiedResetToken).Scan(&userEmail)
	if err != nil {
		return "nil", errors.Wrap(err, "error 'set command to redis' repository/inqurys_redis GetResetTokenForCheckUserEmail()")
	}

	return userEmail, nil
}
