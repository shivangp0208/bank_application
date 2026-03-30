package token

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/shivangp0208/bank_application/util"
)

var config util.Config
var err error

var ErrInvalidSecretKey error = errors.New("invalid secret key")
var ErrExpiredToken error = errors.New("invalid token, token is expired")
var ErrInvalidToken error = errors.New("invalid token")

func init() {
	config, err = util.LoadConfig("../")
	if err != nil {
		log.Fatalf("unable to load configuration from config file: %v", err)
	}
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

func (j *JWTMaker) CreateToken(userName string, duration time.Duration) (string, error) {
	payload, err := NewPayload(userName, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(j.secretKey))
}

func (j *JWTMaker) VerifyToken(token string) (*Payload, error) {

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unable to parse token, invalid token signing signature")
		}
		return []byte(j.secretKey), nil
	}

	parsedToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, errors.New("unable to parse token claims" + err.Error())
	}

	payload, ok := parsedToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}
	return payload, nil
}
