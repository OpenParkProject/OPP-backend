package auth

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var DEBUG_MODE = os.Getenv("DEBUG_MODE")

// URL to the public key set for JWT validation
var AUTH_URL = os.Getenv("AUTH_URL")
var PUBKEY_ENDPOINT = os.Getenv("PUBKEY_ENDPOINT")

// URL to the OTP validation endpoint
var OTP_ENDPOINT = os.Getenv("OTP_ENDPOINT")
var USER_BY_OTP_ENDPOINT = os.Getenv("USER_BY_OTP_ENDPOINT")

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrFailedToGetUser = errors.New("failed to get user from context")
	ErrFailedToGetRole = errors.New("failed to get role from context")
)

// fetchPubKey fetches the public key set from the auth service
func fetchPubKey(url string) (*rsa.PublicKey, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, errors.New("Failed to fetch public key set: " + resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonData map[string]any
	jwtPublicKeySet := ""
	if err := json.Unmarshal(body, &jsonData); err == nil {
		if pubkey, ok := jsonData["pubkey"].(string); ok && pubkey != "" {
			jwtPublicKeySet = pubkey
		}
	}
	if jwtPublicKeySet == "" {
		return nil, errors.New("Empty public key")
	}

	var PublicKey *rsa.PublicKey
	publicKeyPEM, err := base64.StdEncoding.DecodeString(jwtPublicKeySet)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, rsa.ErrDecryption
	}
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	PublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok || PublicKey == nil {
		return nil, errors.New("Failed to parse public key")
	}

	return PublicKey, nil
}

// get and returns the PublicKey for JWT validation
func keyFunc(token *jwt.Token) (any, error) {
	var jwtPublicKeySet *rsa.PublicKey
	var err error

	if DEBUG_MODE == "true" {
		jwtPublicKeySet = nil
	} else {
		jwtPublicKeySet, err = fetchPubKey(AUTH_URL + PUBKEY_ENDPOINT)
		if err != nil {
			fmt.Print("Failed to fetch public key set: " + err.Error())
		}
	}

	if jwtPublicKeySet == nil {
		return nil, errors.New("public key set is not available")
	}

	if token.Method.Alg() != jwt.SigningMethodRS512.Alg() {
		return nil, errors.New("unexpected signing method")
	}

	return jwtPublicKeySet, nil
}

func AuthenticationFunc(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	req := input.RequestValidationInput.Request
	if req == nil {
		return errors.New("missing HTTP request in authentication input")
	}

	// Bypass authentication in debug mode
	if DEBUG_MODE == "true" {
		ctx = context.WithValue(ctx, "username", "superuser_debug")
		ctx = context.WithValue(ctx, "role", "superuser")
		*req = *req.WithContext(ctx)
		return nil
	}

	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return errors.New("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.New("invalid Authorization header format")
	}

	tokenstr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenstr, keyFunc)
	if err != nil {
		return err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return errors.New("invalid token")
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return errors.New("token expired")
		}
	} else {
		return errors.New("invalid or missing expiration claim")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return errors.New("missing username in token claims")
	}
	role, ok := claims["role"].(string)
	if !ok {
		return errors.New("missing role in token claims")
	}

	// Update the request context with the username and role
	ctx = context.WithValue(ctx, "username", username)
	ctx = context.WithValue(ctx, "role", role)
	*req = *req.WithContext(ctx)

	return nil
}

func GetPermissions(c *gin.Context) (string, string, error) {
	// Auth middleware sets values in request context
	// not in gin context
	username := c.Request.Context().Value("username")
	if username == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized.Error()})
		return "", "", ErrUnauthorized
	}
	usernameStr, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrFailedToGetUser.Error()})
		return "", "", ErrFailedToGetUser
	}
	role := c.Request.Context().Value("role")
	if role == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized.Error()})
		return "", "", ErrUnauthorized
	}
	roleStr, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrFailedToGetRole.Error()})
		return "", "", ErrFailedToGetRole
	}

	return usernameStr, roleStr, nil
}

func ValidateOTP(otp string) error {
	if DEBUG_MODE == "true" {
		return nil // Bypass OTP validation in debug mode
	}

	payload := map[string]string{"otp": otp}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", AUTH_URL+OTP_ENDPOINT, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OTP validation failed: %s", resp.Status)
	}

	return nil
}

func GetUsernameFromOTP(otp string) (string, error) {
	if DEBUG_MODE == "true" {
		return "debug_user", nil // Return debug username in debug mode
	}

	// USER_BY_OTP_ENDPOINT = /users/me/{otp}
	userByOTPEndpoint := strings.Replace(USER_BY_OTP_ENDPOINT, "{otp}", otp, 1)
	endpoint := AUTH_URL + userByOTPEndpoint
	fmt.Printf("Getting username with OTP at: %s\n", endpoint)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to get username from OTP: %s - %s", resp.Status, string(body))
	}

	var response struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Username, nil
}
