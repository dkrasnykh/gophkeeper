package clients

import (
	"github.com/gorilla/websocket"
	"sync"
)

type UserWSConnMap struct {
	mu    *sync.RWMutex
	value map[int64][]*websocket.Conn
}

func NewUserWSConnMap() *UserWSConnMap {
	return &UserWSConnMap{
		mu:    &sync.RWMutex{},
		value: make(map[int64][]*websocket.Conn),
	}
}

func (m *UserWSConnMap) Put(userID int64, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.value[userID] = append(m.value[userID], conn)
}

func (m *UserWSConnMap) UserConns(userID int64) []*websocket.Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.value[userID]
}
