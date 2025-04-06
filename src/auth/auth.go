package auth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/golang-jwt/jwt/v5"
)

var DEBUG_MODE = os.Getenv("DEBUG_MODE") == "true"

type CustomClaims struct {
	Permissions []string `json:"permissions"`
	Role        string   `json:"role"`
	jwt.RegisteredClaims
}

type Auth struct {
	jwtPublicKeySet string
}

// URL to the public key set for JWT validation
const PUBKEY_URL = "http://localhost:8080/sessions/jwks.json"

func fetchPubKey(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		panic("Failed to fetch public key set")
	}
	if resp.StatusCode != http.StatusOK {
		panic("Failed to fetch public key set")
	}
	defer resp.Body.Close()
	key, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Failed to read public key set")
	}
	jwtPublicKeySet := string(key)
	if jwtPublicKeySet == "" {
		panic("Public key set is empty")
	}
	return jwtPublicKeySet
}

func NewAuth() *Auth {
	var jwtPublicKeySet string

	if DEBUG_MODE {
		jwtPublicKeySet = ""
	} else {
		jwtPublicKeySet = fetchPubKey(PUBKEY_URL)
	}

	return &Auth{
		jwtPublicKeySet: jwtPublicKeySet,
	}
}

// Key function for JWT validation, uses the public key set from the Auth instance
func (auth *Auth) keyfunct(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, errors.New("unexpected signing method")
	}
	if _, ok := token.Header["kid"].(string); !ok {
		return nil, errors.New("missing key ID in token header")
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(auth.jwtPublicKeySet))
	if err != nil {
		return nil, errors.New("failed to parse public key")
	}
	return key, nil
}

// Authentication function for validating JWT tokens
func (auth *Auth) AuthenticationFunc(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	// Bypass authentication in debug mode
	if DEBUG_MODE {
		return nil
	}

	req := input.RequestValidationInput.Request
	if req == nil {
		return errors.New("missing HTTP request in authentication input")
	}

	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return errors.New("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.New("invalid Authorization header format")
	}

	tokenstr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenstr, &CustomClaims{}, auth.keyfunct)
	if err != nil {
		return errors.New("failed to parse token")
	}
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return errors.New("invalid token")
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return errors.New("token expired")
	}

	// TODO: Add additional permission checks if needed

	return nil
}
