package crypto

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashUserPassword(password string) (string, error) {
	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)

	return string(passHashBytes), err
}

func VerifyUserPassword(attempt, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(password), []byte(attempt))
}

type WlClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

const SessionTokenCookie = "weblens-session-token"

func GenerateJWTCookie(username string) (string, error) {
	expires := time.Now().Add(time.Hour * 24 * 7).In(time.UTC)
	claims := WlClaims{
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("weblens_super_secret_key"))
	if err != nil {
		return "", err
	}

	cookie := fmt.Sprintf("%s=%s;Path=/;Expires=%s;HttpOnly", SessionTokenCookie, signedToken, expires.Format(time.RFC1123))
	return cookie, nil
}
