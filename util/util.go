package util

import (
	"fmt"
	"os"
	"time"
	"worstTwitter/user"

	"github.com/golang-jwt/jwt"
)

var secret = os.Getenv("SECRET")

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
