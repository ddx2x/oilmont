package token

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var salt string

func Decode(tokenStr string) (*CustomClaims, error) {
	token := &token{}
	return token.Decode(tokenStr)
}

func Encode(issuer, userName string, expireTime int64) (string, error) {
	token := &token{}
	return token.Encode(issuer, userName, expireTime)
}

type CustomClaims struct {
	UserName string `json:"userName"`
	jwt.StandardClaims
}

type token struct {
	privateKey []byte
}

func (t *token) Decode(tokenStr string) (*CustomClaims, error) {
	if len(t.privateKey) == 0 {
		t.privateKey = []byte(salt)
	}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return t.privateKey, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

func (t *token) Encode(issuer, userName string, expireTime int64) (string, error) {
	if len(t.privateKey) == 0 {
		t.privateKey = []byte(salt)
	}
	return jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		CustomClaims{
			UserName: userName,
			StandardClaims: jwt.StandardClaims{
				Issuer:    issuer,
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: expireTime,
			},
		}).SignedString(t.privateKey)
}

func init() {
	if _salt := os.Getenv("SLAT"); salt != "" {
		salt = "default-private-key"
	} else {
		salt = _salt
	}
}
