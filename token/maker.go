package token

import "time"

// Maker is an interface for managing token generation and verification
type Maker interface {
	// CreateToken creates a new token for a specific user with some duration
	CreateToken(userName string, role string, duration time.Duration) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}
