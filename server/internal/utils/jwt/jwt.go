package jwtauth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomAccessClaims struct {
	jwt.RegisteredClaims
	UserID uint `json:"userId"`
	Username string `json:"username"`
}

func GenerateAccessToken(userID uint, username string, accessSecret []byte) (string, error) {
	claims := CustomAccessClaims{
		UserID: userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessSecret)
}

type CustomRefreshClaims struct {
	jwt.RegisteredClaims
	UserID       uint `json:"userId"`
	TokenVersion uint `json:"tokenVersion"`
	Username string `json:"username"`
}

func GenerateRefreshToken(userID, tokenVersion uint, username string, refreshSecret []byte) (string, error) {
	claims := CustomRefreshClaims{
		UserID:       userID,
		TokenVersion: tokenVersion,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecret)
}

func GetTokenVersion(token string, refreshSecret []byte) uint {
	tokenClaims := &CustomRefreshClaims{}
	_, err := jwt.ParseWithClaims(token, tokenClaims, func(token *jwt.Token) (any, error) {
		return refreshSecret, nil
	})
	if err != nil {
		return 0
	}
	return tokenClaims.TokenVersion
}

func ValidateAccessToken(token string, accessSecret []byte) (uint, error) {
	tokenClaims := &CustomAccessClaims{}
	_, err := jwt.ParseWithClaims(token, tokenClaims, func(token *jwt.Token) (any, error) {
		return accessSecret, nil
	})
	if err != nil {
		return 0, err
	}

	if tokenClaims.ExpiresAt.Before(time.Now()) {
		return 0, errors.New("token expired")
	}

	return tokenClaims.UserID, nil
}

func ValidateRefreshToken(token string, refreshSecret []byte) (uint, error) {
	tokenClaims := &CustomRefreshClaims{}
	_, err := jwt.ParseWithClaims(token, tokenClaims, func(token *jwt.Token) (any, error) {
		return refreshSecret, nil
	})
	if err != nil {
		return 0, err
	}

	if tokenClaims.ExpiresAt.Before(time.Now()) {
		return 0, errors.New("token expired")
	}

	return tokenClaims.UserID, nil
}
