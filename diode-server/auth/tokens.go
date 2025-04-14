package auth

import "github.com/golang-jwt/jwt/v4"

// TokenParser is an interface for parsing JWT tokens
type TokenParser interface {
	Parse(tokenString string, keyfunc jwt.Keyfunc) (*jwt.Token, error)
}

// JWTParser is a struct that implements the TokenParser interface
type JWTParser struct{}

// Parse parses a JWT token using the provided keyfunc
func (p JWTParser) Parse(tokenString string, keyfunc jwt.Keyfunc) (*jwt.Token, error) {
	return jwt.Parse(tokenString, keyfunc)
}
