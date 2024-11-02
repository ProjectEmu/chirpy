package authy

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GetBearerToken extracts the bearer token from the Authorization header.
func GetBearerToken(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", errors.New("authorization header missing")
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	bearer := strings.TrimPrefix(header, "Bearer ")
	if bearer == "" {
		return "", errors.New("bearer token is empty")
	}

	return bearer, nil
}

// MakeJWT creates and returns a JWT for the given user ID, using the provided secret and expiration duration.
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// Define the claims for the token
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}

	// Create a new token with the specified claims and signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the provided secret key
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	// Define the claims structure
	claims := &jwt.RegisteredClaims{}

	// Parse the token with the claims and the provided secret key
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return uuid.Nil, jwt.ErrSignatureInvalid
	}

	// Parse the user ID from the claims
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// HashPassword takes a plaintext password as input and returns a bcrypt hashed version of it.
func HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword generates a bcrypt hash of the password using a cost factor.
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash compares a plain-text password with a bcrypt hash and returns an error if they do not match.
func CheckPasswordHash(password, hash string) error {
	// bcrypt.CompareHashAndPassword returns nil if the password matches the hash.
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}
