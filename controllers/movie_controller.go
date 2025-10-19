package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	db "github.com/Neph-dev/MovieStreamServer/database"
	"github.com/Neph-dev/MovieStreamServer/models"
	"github.com/Neph-dev/MovieStreamServer/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var movieCollection *mongo.Collection = db.OpenCollection("movies")
var rankingCollection *mongo.Collection = db.OpenCollection("rankings")

func GetMovies() gin.HandlerFunc {
	return func(_context *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movies []models.Movie

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
		var movie models.Movie

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

        var movie models.Movie

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

var reviewUpdate struct {
	AdminReview string `json:"admin_review" validate:"required"`
}
var reviewUpdateResponse struct {
	RankingName string `json:"ranking_name"`
	AdminReview string `json:"admin_review"`
}

func AdminReviewUpdate() gin.HandlerFunc {
	return func(_context *gin.Context) {
		role, exists := _context.Get("role")
		fmt.Println(role, exists)
		if !exists || role != "ADMIN" {
			_context.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admins only"})
			return
		}

		imdbID := _context.Param("imdb_id")

		if err := _context.BindJSON(&reviewUpdate); err != nil {
			_context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		sentiment, rankValue, err := GetReviewRanking(reviewUpdate.AdminReview)
		if err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting review ranking"})
			return
		}
		log.Printf("Determined sentiment: %s with rank value: %d", sentiment, rankValue)

		var validate = validator.New()
		if err := validate.Struct(reviewUpdate); err != nil {
			_context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		exists, err = utils.DocumentExists(ctx, movieCollection, bson.M{"imdb_id": imdbID})
		
		if err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for existing movie"})
			return
		}
		if !exists {
			_context.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		err = utils.UpdateDocument(
			ctx,
			movieCollection,
			bson.M{"imdb_id": imdbID},
			bson.M{"$set": bson.M{
				"admin_review": reviewUpdate.AdminReview, 
				"ranking": bson.M{
					"ranking_name": sentiment,
					"ranking_value": rankValue,
				},
			}},
		)

		if err != nil {
			_context.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating admin review"})
			return
		}

		_context.JSON(http.StatusOK, gin.H{
			"message": "Admin review updated successfully",
			"ranking_name": sentiment,
			"admin_review": reviewUpdate.AdminReview,
		})
	}
}

func GetReviewRanking(admin_review string) (string, int, error) {
	rankings, err := GetRankings()
	if err != nil {
		return "", 0, err
	}

	sentimentDelimited := ""

	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited += ranking.RankingName + ","
		}
	}

	sentimentDelimited = strings.Trim(sentimentDelimited, ",")

	err = godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, proceeding with existing environment variables")
		return "", 0, err
	}

	OpenAIKey := os.Getenv("OPENAI_API_KEY")
	
	if OpenAIKey == "" {
		return "", 0, errors.New("OPENAI_API_KEY not set in environment")
	}

	llm, err := openai.New(openai.WithToken(OpenAIKey))
	if err != nil {
		return "", 0, err
	}

	prompt := strings.Replace(os.Getenv("BASE_PROMPT_TEMPLATE"), "{rankings}", sentimentDelimited, 1)

	response, err := llm.Call(context.Background(), prompt + admin_review)
	if err != nil {
		return "", 0, err
	}
	rankVal := 0

	for _, ranking := range rankings {
		if strings.EqualFold(strings.TrimSpace(response), ranking.RankingName) {
			rankVal = ranking.RankingValue
			break
		}
	}

	if rankVal == 0 {
		return "", 0, errors.New("could not determine ranking from response")
	}

	return response, rankVal, nil
}

func GetRankings() ([]models.Ranking, error) {
	var rankings []models.Ranking

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cursor, err := rankingCollection.Find(ctx, bson.M{})
	if err != nil {
		return rankings, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &rankings); err != nil {
		return rankings, err
	}

	return rankings, nil
}