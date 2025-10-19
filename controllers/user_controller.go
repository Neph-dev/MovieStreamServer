package controllers

import (
	"context"
	"net/http"
	"time"

	db "github.com/Neph-dev/MovieStreamServer/database"
	model "github.com/Neph-dev/MovieStreamServer/models"
	"github.com/Neph-dev/MovieStreamServer/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

var userCollection = db.OpenCollection("users")

func HashPassword(password string) (string, error) {
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(HashPassword), nil
}

func RegisterUser () gin.HandlerFunc {
	return func (_context * gin.Context) {
		var user model.User

		if err := _context.BindJSON(&user); err != nil {
			_context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		var validate = validator.New()

		if err := validate.Struct(user); err != nil {
			_context.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}
		user.Password = hashedPassword

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		if exists, err := utils.DocumentExists(ctx, userCollection, bson.M{"email": user.Email}); err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for existing user"})
			return
		} else if exists {
			_context.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			return
		}

		user.UserID = bson.NewObjectID().Hex()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		if err := utils.InsertDocument(ctx, userCollection, user); err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting user into database"})
			return
		}

		_context.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	}
}



