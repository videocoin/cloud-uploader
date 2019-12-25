package auth

import (
	"fmt"
	"context"
	"errors"

	jwt "github.com/dgrijalva/jwt-go"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

var (
	ErrInvalidSecret = errors.New("invalid secret")
	ErrInvalidToken1  = errors.New("invalid token1")
	ErrInvalidToken2  = errors.New("invalid token2")
)

type TokenType int32

type ExtendedClaims struct {
	Type TokenType `json:"token_type"`
	jwt.StandardClaims
}

func AuthFromContext(ctx context.Context) (context.Context, string, error) {
	secret, ok := SecretKeyFromContext(ctx)
	fmt.Printf("\nsecret %v\n", secret)
	if !ok {
		return ctx, "", ErrInvalidSecret
	}

	jwtToken, err := grpcauth.AuthFromMD(ctx, "bearer")
	fmt.Printf("\n jwtToken %v\n", jwtToken)
	if err != nil {
		fmt.Printf("err %v", err)
		return ctx, "", err
	}

	t, err := jwt.ParseWithClaims(jwtToken, &ExtendedClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return ctx, "", err
	}

	if !t.Valid {
		return ctx, "", ErrInvalidToken2
	}

	ctx = NewContextWithUserID(ctx, t.Claims.(*ExtendedClaims).Subject)
	ctx = NewContextWithType(ctx, t.Claims.(*ExtendedClaims).Type)

	return ctx, jwtToken, nil
}
