package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/shivangp0208/bank_application/util"
)

var config util.Config

var ErrInvalidSecretKey error = errors.New("invalid secret key")
var ErrExpiredToken error = errors.New("invalid token, token is expired")
var ErrInvalidToken error = errors.New("invalid token")

func init() {
	config = util.GetConfig()
}

type JWTMaker struct {
	secretKey string
}

func NewJwtMaker(secretKey string) (jwtMaker Maker, err error) {
	if len(secretKey) < config.MinSecretKeyLength {
		return nil, ErrInvalidSecretKey
	}
	return &JWTMaker{secretKey}, nil
}

func (j *JWTMaker) CreateToken(userName string, role string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(userName, role, duration)
	if err != nil {
		return "", payload, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := jwtToken.SignedString([]byte(j.secretKey))
	return signedToken, payload, err
}

func (j *JWTMaker) VerifyToken(token string) (*Payload, error) {

	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unable to parse token, invalid token signing signature")
		}
		return []byte(j.secretKey), nil
	}

	parsedToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && (verr.Error() == ErrExpiredToken.Error()) {
			return nil, ErrExpiredToken
		}
		return nil, errors.New("unable to parse token claims: " + err.Error())
	}

	payload, ok := parsedToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}
	return payload, nil
}
