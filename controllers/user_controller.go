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

func LoginUser() gin.HandlerFunc {
	var userCollection = db.OpenCollection("users")

	return func(_context *gin.Context) {
		var userLogin model.UserLogin
		if err := _context.BindJSON(&userLogin); err != nil {
			_context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		var validate = validator.New()
		if err := validate.Struct(userLogin); err != nil {
			_context.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user model.User
		err := userCollection.FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(&user)
		if err != nil {
			_context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userLogin.Password))
		if err != nil {
			_context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		token, refreshToken, err := utils.GenerateAllTokens(user.Email, user.FirstName, user.LastName, user.UserID, user.Role)
		if err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
			return
		}

		user.Token = token
		user.RefreshToken = refreshToken

		err = utils.UpdateTokens(token, refreshToken, user.UserID)
		if err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating tokens"})
			return
		}

		_context.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"st-access-token": user.Token,
			"st-refresh-token": user.RefreshToken,
		})
		// _context.SetCookie("token", user.Token, 3600, "/", "localhost", false, true)
		// _context.SetCookie("refresh_token", user.RefreshToken, 3600, "/", "localhost", false, true)
	}
}



