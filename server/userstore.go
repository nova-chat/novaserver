package server

import (
	"sync"

	"github.com/google/uuid"
	"github.com/nova-chat/novaproto"
)

type Client struct {
	Id uuid.UUID

	WireStream   *novaproto.NovaWireStreamCipher
	PacketStream *novaproto.RoutedPacketStream
}

type userStore struct {
	Users    map[uuid.UUID]*Client
	usersMut sync.RWMutex
}

func newUserStore() *userStore {
	return &userStore{
		Users: make(map[uuid.UUID]*Client),
	}
}

func (s *userStore) SetUser(u *Client) {
	s.usersMut.Lock()
	defer s.usersMut.Unlock()
	s.Users[u.Id] = u
}

func (s *userStore) DelUser(id uuid.UUID) {
	s.usersMut.Lock()
	defer s.usersMut.Unlock()
	delete(s.Users, id)
}

func (s *userStore) GetUser(id uuid.UUID) (*Client, bool) {
	s.usersMut.Lock()
	defer s.usersMut.Unlock()
	u, ok := s.Users[id]
	return u, ok
}
