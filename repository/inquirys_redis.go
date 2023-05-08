package repository

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type RedisClient struct {
	redisDb *redis.Client
}

func NewRedisreposiory(redis *redis.Client) *RedisClient {
	return &RedisClient{redisDb: redis}
}

type RedisInqurysInterface interface {
	AddAccessToken(ctx context.Context, usermail, accessToken string) (string, error)
	GetAccessTokenByUsermail(ctx context.Context, usermail string) (string, error)
}

// вставить access token  и username
func (r *RedisClient) AddAccessToken(ctx context.Context, usermail, accessToken string) (string, error) {
	err := r.redisDb.Set(ctx, usermail, accessToken, time.Minute*15).Err()
	if err != nil {
		errors.Wrapf(err, "error 'set comand to redis' repository/inqurys_redis  AddAccessToken()")
	}
	return "ok", err
}

func (r *RedisClient) GetAccessTokenByUsermail(ctx context.Context, usermail string) (string, error) {
	var acceessToken string
	err := r.redisDb.Get(ctx, usermail).Scan(&acceessToken)
	if err != nil {
		errors.Wrapf(err, "error 'get comand to redis' repository/inqurys_redis  GetAccessTokenByUsermail()")
	}
	return acceessToken, err
}
