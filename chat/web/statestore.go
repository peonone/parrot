package web

import (
	"github.com/go-redis/redis"
	web "github.com/micro/go-web"
	"github.com/peonone/parrot/chat"
)

type stateStore interface {
	online(uid string) error
	offline(uid string) error
}

type rdsStateStore struct {
	*redis.Client
}

func (s *rdsStateStore) online(uid string) error {
	_, err := s.HSet(chat.OnlineStateKey, uid, web.DefaultId).Result()
	return err
}

func (s *rdsStateStore) offline(uid string) error {
	_, err := s.HDel(chat.OnlineStateKey, uid).Result()
	return err
}
