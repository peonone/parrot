package srv

import "github.com/go-redis/redis"

const TokenKey = "auth-token"

type tokenStore interface {
	saveToken(uid string, token string) error
	getToken(uid string) (string, error)
}

type redisTokenStore struct {
	*redis.Client
}

func (r *redisTokenStore) saveToken(uid string, token string) error {
	cmd := r.HSet(TokenKey, uid, token)
	_, err := cmd.Result()
	return err
}

func (r *redisTokenStore) getToken(uid string) (string, error) {
	return r.HGet(TokenKey, uid).Result()
}
