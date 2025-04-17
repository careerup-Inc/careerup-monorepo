package repository

import (
	"context"
	"errors"
	"time"

	"github.com/careerup-Inc/careerup-monorepo/services/avatar-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AvatarRepository struct {
	collection *mongo.Collection
}

func NewAvatarRepository(db *mongo.Database) *AvatarRepository {
	return &AvatarRepository{
		collection: db.Collection("avatars"),
	}
}

func (r *AvatarRepository) Create(ctx context.Context, avatar *model.Avatar) error {
	avatar.CreatedAt = time.Now()
	avatar.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, avatar)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		avatar.ID = oid.Hex()
	}

	return nil
}

func (r *AvatarRepository) GetByID(ctx context.Context, id string) (*model.Avatar, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	var avatar model.Avatar
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&avatar)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("avatar not found")
		}
		return nil, err
	}

	return &avatar, nil
}

func (r *AvatarRepository) Update(ctx context.Context, id string, update *model.AvatarUpdateRequest) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	updateDoc := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if update.Style != "" {
		updateDoc["$set"].(bson.M)["style"] = update.Style
	}
	if update.Features != nil {
		updateDoc["$set"].(bson.M)["features"] = update.Features
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": oid}, updateDoc)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("avatar not found")
	}

	return nil
}

func (r *AvatarRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("avatar not found")
	}

	return nil
}
