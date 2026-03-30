package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, err
}

func (p *Payload) Valid() error {
	if len(p.Username) == 0 {
		return errors.New("invalid payload username, cannot be empty")
	}
	if p.ID == uuid.Nil {
		return errors.New("invalid payload ID, cannot be empty")
	}
	if time.Now().After(p.ExpiredAt) {
		return errors.New("invalid payload, token expired")
	}
	return nil
}
