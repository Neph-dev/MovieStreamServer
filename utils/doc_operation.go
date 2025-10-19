package utils

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InsertDocument(ctx context.Context, collection *mongo.Collection, document interface{}) error {
	_, err := collection.InsertOne(ctx, document)
	return err
}

func DocumentExists(ctx context.Context, collection *mongo.Collection, filter bson.M) (bool, error) {
	count, err := collection.CountDocuments(ctx, filter)
	
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, err
}