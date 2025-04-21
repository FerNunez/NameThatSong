package dbstore

import (
	"fmt"
	"github.com/google/uuid"

	"github.com/FerNunez/NameThatSong/internal/store"
)

type SessionStore struct {
	db map[string]store.Session
}

func (s *SessionStore) create(session *store.Session) error {
	s.db[session.SessionID] = *session
	return nil
}
func (s *SessionStore) fetch(sessionID string) (*store.Session, error) {
	val, ok := s.db[sessionID]
	if !ok {
		return nil, fmt.Errorf("no user associated with the session")
	}

	return &val, nil
}

// type NewSessionStoreParams struct {
// 	DB SessionStore
// }

func NewSessionStore() *SessionStore {
	return &SessionStore{
		db: make(map[string]store.Session),
	}
}

func (s *SessionStore) CreateSession(session *store.Session) (*store.Session, error) {
	session.SessionID = uuid.New().String()

	err := s.create(session)
	return session, err
}

func (s *SessionStore) GetUserFromSession(sessionID string) (*store.Session, error) {
	session, err := s.fetch(sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}
