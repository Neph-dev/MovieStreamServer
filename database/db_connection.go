package database

import (
	"fmt"
	"log"
	"os"

	dotenv "github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectDB() *mongo.Client {
	err := dotenv.Load()
	
	if err != nil {
		log.Fatal("Error: Error loading .env file")
	}

	MongoDb := os.Getenv("MONGO_URI")

	if MongoDb == "" {
		log.Fatal("Error: MONGO_URI not found in environment variables")
	}

	clientOptions := options.Client().ApplyURI(MongoDb)

	client, err := mongo.Connect(nil, clientOptions)
	if err != nil {
		log.Fatal("Error: Could not connect to MongoDB:", err)
	}

	fmt.Println("Connected to MongoDB!")
	return client
}

func OpenCollection(collectionName string) *mongo.Collection {
	err := dotenv.Load()
	if err != nil {
		log.Fatal("Error: Error loading .env file")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("Error: DB_NAME not found in environment variables")
	}

	client := ConnectDB()

	collection := client.Database(dbName).Collection(collectionName)
	if collection == nil {
		log.Fatalf("Error: Could not open collection %s", collectionName)
	}
	
	return collection
}