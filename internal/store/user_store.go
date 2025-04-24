package store

import (
	"context"
	"time"

	"github.com/FerNunez/NameThatSong/internal/store/database"
	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Email          string
	HashedPassword string
}

type UserStore interface {
	Create(ctx context.Context, email, hashed_password string) (User, error)
	GetEmail(ctx context.Context, email string) (User, error)
	UpdateById(ctx context.Context, id uuid.UUID, newEmail, newHashedPass string) error
	Reset(ctx context.Context) error
}

type SQLUserStore struct {
	db *database.Queries
}

func NewSQLUserStore(db *database.Queries) UserStore {
	return &SQLUserStore{
		db: db,
	}
}

func (s *SQLUserStore) Create(ctx context.Context, email, hashed_password string) (User, error) {
	dbUser, err := s.db.CreateUser(ctx, database.CreateUserParams{
		Email:          email,
		HashedPassword: hashed_password,
	})
	if err != nil {
		return User{}, err
	}

	return User{
		ID:             dbUser.ID,
		CreatedAt:      dbUser.CreatedAt,
		UpdatedAt:      dbUser.UpdatedAt,
		Email:          dbUser.Email,
		HashedPassword: dbUser.HashedPassword,
	}, nil
}

func (s SQLUserStore) GetEmail(ctx context.Context, email string) (User, error) {
	dbUser, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:             dbUser.ID,
		CreatedAt:      dbUser.CreatedAt,
		UpdatedAt:      dbUser.UpdatedAt,
		Email:          dbUser.Email,
		HashedPassword: dbUser.HashedPassword,
	}, nil
}

func (s *SQLUserStore) UpdateById(ctx context.Context, id uuid.UUID, newEmail, newHashedPass string) error {
	return s.db.UpdateUserLoginByID(ctx, database.UpdateUserLoginByIDParams{
		Email:          newEmail,
		HashedPassword: newHashedPass,
		ID:             id,
	})
}
func (s *SQLUserStore) Reset(ctx context.Context) error {

	return s.db.ResetUsers(ctx)

}
