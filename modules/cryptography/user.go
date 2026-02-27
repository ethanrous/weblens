package cryptography

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// BcryptDifficultyCtxKey is the context key for specifying bcrypt difficulty level.
const BcryptDifficultyCtxKey = "bcryptDifficulty"
const bcryptDefaultDifficulty = 11

// HashUserPasswordDifficulty hashes a user password using bcrypt with the specified difficulty level.
func HashUserPasswordDifficulty(password string, difficulty int) ([]byte, error) {
	difficulty = max(difficulty, bcrypt.MinCost)

	return bcrypt.GenerateFromPassword([]byte(password), difficulty)
}

// HashUserPassword hashes a user password using bcrypt.
func HashUserPassword(ctx context.Context, password string) (string, error) {
	// For testing, we can set the bcrypt difficulty in the context
	bcryptDifficultyI := ctx.Value(BcryptDifficultyCtxKey)

	bcryptDifficulty, ok := bcryptDifficultyI.(int)
	if !ok {
		bcryptDifficulty = bcryptDefaultDifficulty
	}

	passHashBytes, err := HashUserPasswordDifficulty(password, bcryptDifficulty)
	if err != nil {
		return "", err
	}

	return string(passHashBytes), err
}

// VerifyUserPassword verifies that an attempted password matches the hashed password.
func VerifyUserPassword(attempt, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(password), []byte(attempt))
}

// WlClaims represents the JWT claims for Weblens authentication tokens.
type WlClaims struct {
	jwt.RegisteredClaims

	Username string `json:"username"`
}

// SessionTokenCookie is the name of the HTTP cookie that stores the session token.
const SessionTokenCookie = "weblens-session-token"

// UserCrumbCookie is the name of the HTTP cookie that stores the username.
const UserCrumbCookie = "weblens-user-name"

var superSecretKey = []byte("weblens_super_secret_key")

// GenerateJWT generates a JWT token for the specified username.
func GenerateJWT(username string) (string, time.Time, error) {
	expires := time.Now().Add(time.Hour * 24 * 7).In(time.UTC)
	claims := WlClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
		},

		username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(superSecretKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expires, nil
}

// GetUsernameFromToken extracts and validates the username from a JWT token string.
func GetUsernameFromToken(tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", wlerrors.New("no jwt provided")
	}

	jwtToken, err := jwt.ParseWithClaims(
		tokenStr,
		&WlClaims{},
		func(_ *jwt.Token) (any, error) {
			return superSecretKey, nil
		},
	)
	if err != nil {
		if wlerrors.Is(err, jwt.ErrTokenExpired) {
			return "", wlerrors.New("jwt expired")
		}

		return "", wlerrors.WithStack(err)
	}

	username := jwtToken.Claims.(*WlClaims).Username

	return username, nil
}
