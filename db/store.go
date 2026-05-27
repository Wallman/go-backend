package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const maxRetries = 5

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
		db:      db,
	}
}

// AddTokensUsed increments the token count for the given user using optimistic
// locking. If a concurrent update changes the version before we can commit, the
// operation is retried up to maxRetries times.
func (s *Store) AddTokensUsed(ctx context.Context, id uuid.UUID, tokens int32) (User, error) {
	usr, err := s.Queries.GetUser(ctx, id)
	if err != nil {
		return User{}, err
	}
	for i := 0; i < maxRetries; i++ {
		updated, err := s.Queries.AddTokensUsed(ctx, AddTokensUsedParams{
			ID:         id,
			TokensUsed: tokens,
			Version:    usr.Version,
		})
		if err == nil {
			return updated, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return User{}, err
		}
		// ErrNoRows indicates either a version conflict or the user no longer
		// exists. Re-fetch to distinguish between the two.
		usr, err = s.Queries.GetUser(ctx, id)
		if err != nil {
			return User{}, err
		}
	}
	return User{}, fmt.Errorf("optimistic locking for user %s failed after %d retries due to high contention", id, maxRetries)
}
