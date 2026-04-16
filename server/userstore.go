package server

import (
	"sync"

	"github.com/google/uuid"
	"github.com/nova-chat/novaproto"
)

type User struct {
	Id            uuid.UUID
	EncryptionKey []byte

	WireStream   novaproto.Wire
	PacketStream *novaproto.RoutedPacketStream
}

type userStore struct {
	Users    map[uuid.UUID]*User
	usersMut sync.RWMutex
}

func newUserStore() *userStore {
	return &userStore{
		Users: make(map[uuid.UUID]*User),
	}
}

func (s *userStore) SetUser(u *User) {
	s.usersMut.Lock()
	defer s.usersMut.Unlock()
	s.Users[u.Id] = u
}

func (s *userStore) DelUser(id uuid.UUID) {
	s.usersMut.Lock()
	defer s.usersMut.Unlock()
	delete(s.Users, id)
}

func (s *userStore) GetUser(id uuid.UUID) (*User, bool) {
	s.usersMut.Lock()
	defer s.usersMut.Unlock()
	u, ok := s.Users[id]
	return u, ok
}
