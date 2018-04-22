package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateManager(t *testing.T) {
	mnger := newOnlineUsersManager()

	user1 := &onlineUser{
		userClient: new(mockUserClient),
		uid:        "user1",
		pushCh:     make(chan interface{}),
	}
	user2 := &onlineUser{
		userClient: new(mockUserClient),
		uid:        "user2",
		pushCh:     make(chan interface{}),
	}

	assert.Nil(t, mnger.getOnlineClient(user1.uid))
	mnger.online(user1)
	assert.Equal(t, mnger.getOnlineClient(user1.uid), user1)
	assert.Nil(t, mnger.getOnlineClient(user2.uid))

	mnger.online(user2)
	assert.Equal(t, mnger.getOnlineClient(user2.uid), user2)

	mnger.offline(user1)
	assert.Nil(t, mnger.getOnlineClient(user1.uid))

	mnger.clear()
	assert.Nil(t, mnger.getOnlineClient(user1.uid))
	assert.Nil(t, mnger.getOnlineClient(user2.uid))
}
