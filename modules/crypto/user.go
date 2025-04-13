package crypto

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
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

var superSecretKey = []byte("weblens_super_secret_key")

func GenerateJWT(username string) (string, time.Time, error) {
	expires := time.Now().Add(time.Hour * 24 * 7).In(time.UTC)
	claims := WlClaims{
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(superSecretKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expires, nil
}

func GetUsernameFromToken(tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", errors.New("no jwt provided")
	}

	jwtToken, err := jwt.ParseWithClaims(
		tokenStr,
		&WlClaims{},
		func(token *jwt.Token) (any, error) {
			return superSecretKey, nil
		},
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", errors.New("jwt expired")
		}
		return "", errors.WithStack(err)
	}

	username := jwtToken.Claims.(*WlClaims).Username
	return username, nil
}
