package cryptography

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// BcryptDifficultyCtxKey is the context key for specifying bcrypt difficulty level.
const BcryptDifficultyCtxKey = "bcryptDifficulty"
const bcryptDefaultDifficulty = 11

// HashUserPassword hashes a user password using bcrypt.
func HashUserPassword(ctx context.Context, password string) (string, error) {
	// For testing, we can set the bcrypt difficulty in the context
	bcryptDifficultyI := ctx.Value(BcryptDifficultyCtxKey)

	bcryptDifficulty, ok := bcryptDifficultyI.(int)
	if !ok {
		bcryptDifficulty = bcryptDefaultDifficulty
	}

	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptDifficulty)

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

var jwtSigningKey []byte

func init() {
	ReloadJWTKey()
}

// ReloadJWTKey loads the JWT signing key from the WEBLENS_JWT_SECRET environment
// variable. If the variable is not set, a random 32-byte key is generated. This
// means sessions will not survive server restarts unless the env var is configured.
func ReloadJWTKey() {
	if secret := os.Getenv("WEBLENS_JWT_SECRET"); secret != "" {
		jwtSigningKey = []byte(secret)
	} else {
		key, err := RandomBytes(32)
		if err != nil {
			panic("failed to generate random JWT key: " + err.Error())
		}

		jwtSigningKey = key
	}
}

// ValidateFilename checks that a filename does not contain path traversal
// characters or other invalid sequences that could escape the intended directory.
func ValidateFilename(name string) error {
	if name == "" {
		return wlerrors.New("filename must not be empty")
	}

	if name == "." || name == ".." {
		return wlerrors.New("filename must not be '.' or '..'")
	}

	if strings.ContainsAny(name, "/\\") {
		return wlerrors.New("filename must not contain path separators")
	}

	if len(name) > 255 {
		return wlerrors.New("filename must not exceed 255 characters")
	}

	return nil
}

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

	signedToken, err := token.SignedString(jwtSigningKey)
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
			return jwtSigningKey, nil
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
