package controllers

import (
	"context"
	"net/http"
	"time"

	db "github.com/Neph-dev/MovieStreamServer/database"
	model "github.com/Neph-dev/MovieStreamServer/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var movieCollection *mongo.Collection = db.OpenCollection("movies")

func GetMovies() gin.HandlerFunc {
	return func(_context *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movies []model.Movie

		cursor, err := movieCollection.Find(ctx, bson.M{})
		if err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching movies from database"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &movies); err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding movies from database"})
			return
		}

		_context.JSON(http.StatusOK, movies)
	}
}