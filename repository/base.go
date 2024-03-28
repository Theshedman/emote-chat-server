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
type _[T identifier] interface {
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

func (m *Model[T]) Create(ctx context.Context, entity T) (*T, error) {
	// Dynamic data like ID and Timestamps shouldn't necessarily be generated on DB.
	// Another part of the systems that need access to these dynamic data should have
	// no need to wait for the DB operations to complete before making use of them
	entity.SetID(primitive.NewObjectID()) // Ensure ID generation
	entity.SetTimestamp()                 // Timestamp generation

	_, err := m.collection.InsertOne(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	return &entity, nil
}

func (m *Model[T]) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid object ID: %w", err)
	}

	_, err = m.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	return nil
}

func (m *Model[T]) FindById(ctx context.Context, id string) (*T, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}

	return m.FindOne(ctx, bson.M{"_id": objID})
}

func (m *Model[T]) FindOne(ctx context.Context, filter interface{}) (*T, error) {
	result := m.collection.FindOne(ctx, filter)
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

func (m *Model[T]) Find(ctx context.Context, filter interface{}, page int, limit int, sortKey string) ([]*T, error) {
	findOptions := options.Find()
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))

	if sortKey != "" {
		findOptions.SetSort(bson.D{{sortKey, -1}})
	}

	if filter == nil {
		filter = bson.M{}
	}

	cursor, err := m.collection.Find(ctx, filter, findOptions)
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

func (m *Model[T]) Update(ctx context.Context, id string, entity T) (*T, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}

	update := bson.M{"$set": entity}
	result, err := m.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("entity not found")
	}

	return &entity, nil
}
