package srv

import (
	"github.com/go-redis/redis"
	"github.com/peonone/parrot/chat"
)

type stateStore interface {
	online(uid string, webNode string) error
	offline(uid string) error
	getOnlineWebNode(uid string) (string, error)
}

type redisStateStore struct {
	rdsCli *redis.Client
}

func (m *redisStateStore) online(uid string, webNode string) error {
	_, err := m.rdsCli.HSet(chat.OnlineStateKey, uid, webNode).Result()
	return err
}

func (m *redisStateStore) offline(uid string) error {
	_, err := m.rdsCli.HDel(chat.OnlineStateKey, uid).Result()
	return err
}

func (m *redisStateStore) getOnlineWebNode(uid string) (string, error) {
	return m.rdsCli.HGet(chat.OnlineStateKey, uid).Result()
}
