package utils

import (
	"context"
	"errors"
	"os"
	"time"

	db "github.com/Neph-dev/MovieStreamServer/database"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type SignedDetails struct {
	Email    string
	FirstName string
	LastName  string
	UID      string
	Role     string
	jwt.RegisteredClaims
}

var JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
var JWT_REFRESH_KEY = os.Getenv("JWT_REFRESH_KEY")
var userCollection = db.OpenCollection("users")

func GenerateAllTokens(email string, firstName string, lastName string, UID string, role string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UID:       UID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "MovieStream",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	refreshClaims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UID:       UID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "MovieStream",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 168)), // 7 days
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(JWT_SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(JWT_REFRESH_KEY))
	if err != nil {
		return "", "", err
	}

	err = UpdateTokens(token, refreshToken, UID)
	if err != nil {
		return "", "", err
	}
	
	return token, refreshToken, nil
}

func UpdateTokens(signedToken string, signedRefreshToken string, UID string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	err := UpdateDocument(
		ctx,
		userCollection,
		bson.M{"user_id": UID},
		bson.M{"$set": bson.M{"token": signedToken, "refresh_token": signedRefreshToken, "updated_at": time.Now()}},
	)

	return err
}

func GetAccessToken (_context *gin.Context) (string, error) {
	authHeader := _context.Request.Header.Get("Authorization")
	
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	tokenString := authHeader[len("Bearer "):]
	if tokenString == "" {
		return "", errors.New("bearer token missing")
	}

	return tokenString, nil
}

func ValidateToken(signedToken string) (*SignedDetails, error) {
	claims := &SignedDetails{}

	token, err := jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT_SECRET_KEY), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func GetDataFromContext(_context *gin.Context, field string) (string, error) {
	value, exists := _context.Get(field)
	if !exists {
		return "", errors.New("required data not found in context")
	}
	return value.(string), nil
}