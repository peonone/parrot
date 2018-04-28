package web

import (
	"sync"
)

type onlineUsersManager struct {
	mu         sync.Mutex
	onlineUIDs map[string]*onlineUser
}

func newOnlineUsersManager() *onlineUsersManager {
	return &onlineUsersManager{
		onlineUIDs: make(map[string]*onlineUser),
	}
}

func (m *onlineUsersManager) online(c *onlineUser) {
	m.mu.Lock()
	m.onlineUIDs[c.uid] = c
	m.mu.Unlock()
}

func (m *onlineUsersManager) offline(c *onlineUser) {
	if c.uid != "" {
		m.mu.Lock()
		delete(m.onlineUIDs, c.uid)
		m.mu.Unlock()
	}
	c.mu.Lock()
	c.closed = true
	close(c.pushCh)
	c.mu.Unlock()
}

func (m *onlineUsersManager) clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for uid := range m.onlineUIDs {
		delete(m.onlineUIDs, uid)
	}
}

func (m *onlineUsersManager) getOnlineClient(uid string) *onlineUser {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.onlineUIDs[uid]
}

func (m *onlineUsersManager) getAllOnlineUsers() map[string]*onlineUser {
	return m.onlineUIDs
}
