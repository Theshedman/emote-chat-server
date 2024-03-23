package repository

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Should not be exported outside the package.
type identifier interface {
	GetID() primitive.ObjectID
	SetID(primitive.ObjectID)
	SetTimestamp()
}

// Should not be exported outside the package.
type repository[T identifier] interface {
	Delete(ctx context.Context, id string) error
	Create(ctx context.Context, entity T) (*T, error)
	FindById(ctx context.Context, id string) (*T, error)
	FindOne(ctx context.Context, filter interface{}) (*T, error)
	Find(ctx context.Context, page int, limit int) ([]*T, error)
	Update(ctx context.Context, id string, entity T) (*T, error)
}

type Model[T identifier] struct {
	collection *mongo.Collection
}

// Should not be exported outside the package.
func newModel[T identifier](collection *mongo.Collection) *Model[T] {
	return &Model[T]{collection: collection}
}

func (r *Model[T]) Create(ctx context.Context, entity T) (*T, error) {
	entity.SetID(primitive.NewObjectID()) // Ensure ID generation
	entity.SetTimestamp()

	_, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	return &entity, nil
}

func (r *Model[T]) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid object ID: %w", err)
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	return nil
}

func (r *Model[T]) FindById(ctx context.Context, id string) (*T, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}

	result := r.collection.FindOne(ctx, bson.M{"_id": objID})
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("entity not found")
	} else if result.Err() != nil {
		return nil, fmt.Errorf("failed to find entity: %w", result.Err())
	}

	var entity T
	err = result.Decode(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to decode entity: %w", err)
	}

	return &entity, nil
}

func (r *Model[T]) FindOne(ctx context.Context, filter interface{}) (*T, error) {
	result := r.collection.FindOne(ctx, filter)
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("entity not found")
	} else if result.Err() != nil {
		return nil, fmt.Errorf("failed to find entity: %w", result.Err())
	}

	var entity T
	err := result.Decode(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to decode entity: %w", err)
	}

	return &entity, nil
}

func (r *Model[T]) Find(ctx context.Context, page int, limit int) ([]*T, error) {
	findOptions := options.Find()
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("error finding entities: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, ctx)

	var entities []*T
	if err = cursor.All(ctx, &entities); err != nil {
		return nil, fmt.Errorf("error decoding entities: %w", err)
	}

	return entities, nil
}

func (r *Model[T]) Update(ctx context.Context, id string, entity T) (*T, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}

	update := bson.M{"$set": entity}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("entity not found")
	}

	return &entity, nil
}
