package util

import (
	"crypto/rand"
	"fmt"
	"os"
	"time"
	"worstTwitter/user"

	"github.com/golang-jwt/jwt"
)

var secret string

func init() {
	secret = os.Getenv("SECRET")
	if secret == "" {
		var err error
		secret, err = MakeRandomStr(20)
		if err != nil {
			panic(err)
		}
	}
}

type Claim struct {
	jwt.StandardClaims
	UserID user.UserID
}

func Sign(id user.UserID) string {
	token := jwt.New(jwt.SigningMethodHS256)
	expire := time.Now().Add(time.Hour * 24 * 3)
	token.Claims = &Claim{
		jwt.StandardClaims{
			ExpiresAt: expire.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		id,
	}
	ret, _ := token.SignedString([]byte(secret))
	return ret
}

func UseSession(tokenString string) (user.UserID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid session")
	}
	if claim, ok := token.Claims.(*Claim); ok {
		return claim.UserID, nil
	} else {
		return 0, fmt.Errorf("invalid session")
	}
}

func MakeRandomStr(digit uint32) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, digit)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("unexpected error")
	}

	var result string
	for _, v := range b {
		result += string(letters[int(v)%len(letters)])
	}
	return result, nil
}
