package jwt
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	UserID int64     `json:"user_id"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *Manager) GenerateTokenPair(userID int64, role string) (string, string, error) {
	access, err := m.newToken(userID, role, TokenTypeAccess, m.accessTTL)
	if err != nil {
		return "", "", err
	}
	refresh, err := m.newToken(userID, role, TokenTypeRefresh, m.refreshTTL)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (m *Manager) ParseAccessToken(raw string) (*Claims, error) {
	return m.parse(raw, TokenTypeAccess)
}

func (m *Manager) ParseRefreshToken(raw string) (*Claims, error) {
	return m.parse(raw, TokenTypeRefresh)
}

func (m *Manager) newToken(userID int64, role string, tokenType TokenType, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) parse(raw string, expected TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	if claims.Type != expected {
		return nil, fmt.Errorf("invalid token type")
	}
	return claims, nil
}
































}	return claims, nil	}		return nil, fmt.Errorf("invalid token type")	if claims.Type != expected {	}		return nil, jwt.ErrTokenInvalidClaims	if !ok || !token.Valid {	claims, ok := token.Claims.(*Claims)	}		return nil, err	if err != nil {	})		return m.secret, nil		}			return nil, fmt.Errorf("unexpected signing method")		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (interface{}, error) {func (m *Manager) parse(raw string, expected TokenType) (*Claims, error) {}	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)	}		},			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),			IssuedAt:  jwt.NewNumericDate(time.Now()),		RegisteredClaims: jwt.RegisteredClaims{		Type:   tokenType,		Role:   role,		UserID: userID,	claims := Claims{func (m *Manager) newToken(userID int64, role string, tokenType TokenType, ttl time.Duration) (string, error) {}	return m.parse(raw, TokenTypeRefresh)func (m *Manager) ParseRefreshToken(raw string) (*Claims, error) {}	return m.parse(raw, TokenTypeAccess)func (m *Manager) ParseAccessToken(raw string) (*Claims, error) {}	return access, refresh, nil	}		return "", "", err	if err != nil {	refresh, err := m.newToken(userID, role, TokenTypeRefresh, m.refreshTTL)	}		return "", "", err	if err != nil {	access, err := m.newToken(userID, role, TokenTypeAccess, m.accessTTL)func (m *Manager) GenerateTokenPair(userID int64, role string) (string, string, error) {}	}		refreshTTL: refreshTTL,		accessTTL:  accessTTL,		secret:     []byte(secret),	return &Manager{func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {}	refreshTTL time.Duration	accessTTL  time.Duration	secret     []bytetype Manager struct {}	jwt.RegisteredClaims	Type   TokenType `json:"type"`	Role   string    `json:"role"`	UserID int64     `json:"user_id"`type Claims struct {)	TokenTypeRefresh TokenType = "refresh"	TokenTypeAccess  TokenType = "access"const (type TokenType string)