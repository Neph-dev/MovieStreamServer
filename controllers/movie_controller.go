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

func GetMovieByImdbID() gin.HandlerFunc {
	return func(_context *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		imdbID := _context.Param("imdb_id")
		var movie model.Movie

		err := movieCollection.FindOne(ctx, bson.M{"imdb_id": imdbID}).Decode(&movie)
		if err != nil {
			_context.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		_context.JSON(http.StatusOK, movie)
	}
}

func AddMovie() gin.HandlerFunc {
    return func(_context *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()

        var movie model.Movie

        if err := _context.BindJSON(&movie); err != nil {
            _context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
            return
        }
		
		var validate = validator.New()

		if err := validate.Struct(movie); err != nil {
			_context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

        if exists, err := utils.DocumentExists(ctx, movieCollection, bson.M{"imdb_id": movie.ImdbID}); err != nil {
            _context.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for existing movie"})
            return
        } else if exists {
            _context.JSON(http.StatusConflict, gin.H{"error": "Movie with this IMDB ID already exists"})
            return
        }

        if err := utils.InsertDocument(ctx, movieCollection, movie); err != nil {
            _context.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting movie into database"})
            return
        }

        _context.JSON(http.StatusCreated, movie)
    }
}