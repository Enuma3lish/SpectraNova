package jwt

import (
	"fmt"
	"strconv"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
	jwtv5.RegisteredClaims
}

func GenerateToken(secret string, userID uint64, role string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(secret string, userID uint64, expiry time.Duration) (string, error) {
	claims := jwtv5.RegisteredClaims{
		Subject:   formatUint64(userID),
		ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(expiry)),
		IssuedAt:  jwtv5.NewNumericDate(time.Now()),
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(secret string, tokenString string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenString, &Claims{}, func(token *jwtv5.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwtv5.ErrTokenInvalidClaims
}

func ParseRefreshToken(secret string, tokenString string) (uint64, error) {
	token, err := jwtv5.ParseWithClaims(tokenString, &jwtv5.RegisteredClaims{}, func(token *jwtv5.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*jwtv5.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, jwtv5.ErrTokenInvalidClaims
	}
	return parseUint64(claims.Subject)
}

func formatUint64(v uint64) string {
	return fmt.Sprintf("%d", v)
}

func parseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
