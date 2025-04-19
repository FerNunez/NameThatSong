package dbstore

import (
	"fmt"
	"github.com/google/uuid"

	"github.com/FerNunez/NameThatSong/internal/store"
)

type SessionStore struct {
	db map[string]store.Session
}

func (s *SessionStore) Create(session *store.Session) error {
	s.db[session.SessionID] = *session
	return nil
}
func (s *SessionStore) Fetch(sessionID string) (*store.Session, error) {
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

	err := s.Create(session)
	return session, err
}

func (s *SessionStore) GetUserFromSession(sessionID string) (*store.User, error) {
	session, err := s.Fetch(sessionID)
	if err != nil {
		return nil, err
	}

	return &session.User, nil
}
