package mock

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-pkg/auth"
)

const UserID = "12b1876f-341f-41b0-833f-5312f1e9c308"
const StreamID = "cdc1816b-0be8-44a6-80c3-3e43fbd441ee"

func GetAuthToken(authTokenSecret string) string {
	claims := auth.ExtendedClaims{
		Type: auth.TokenType(usersv1.TokenTypeRegular),
		StandardClaims: jwt.StandardClaims{
			Subject:   UserID,
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	st, err := token.SignedString([]byte(authTokenSecret))
	if err != nil {
		return ""
	}

	return st
}
