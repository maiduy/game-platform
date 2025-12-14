package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	pool "game-platform/internal/platform/pools"
	"game-platform/internal/services/ability/domain"
)

const (
	effectsCollectionName = "effect"
)

// Repository defines the interface for effect data access
type Repository interface {
	CreateEffect(ctx context.Context, effect *domain.Effect) error
	GetEffectByID(ctx context.Context, effectID string) (*domain.Effect, error)
	UpdateEffect(ctx context.Context, effectID string, effect *domain.Effect) error
	DeleteEffect(ctx context.Context, effectID string) error
	ListEffects(ctx context.Context, filter bson.M, page, limit int) ([]domain.Effect, int64, error)
	EffectExists(ctx context.Context, effectID string) (bool, error)
}

// repository implements Repository interface using MongoDB
type repository struct {
	mongoConn *pool.MongoDBConnection
}

// NewRepository creates a new effect repository
func NewRepository(mongoConn *pool.MongoDBConnection) Repository {
	return &repository{
		mongoConn: mongoConn,
	}
}

// getCollection returns the effects collection
func (r *repository) getCollection() *mongo.Collection {
	return r.mongoConn.GetDefaultCollection(effectsCollectionName)
}

// CreateEffect creates a new effect in the database
func (r *repository) CreateEffect(ctx context.Context, effect *domain.Effect) error {
	collection := r.getCollection()

	if effect.ID.IsZero() {
		effect.ID = primitive.NewObjectID()
	}

	_, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		result, err := collection.InsertOne(ctx, effect)
		if err != nil {
			return nil, fmt.Errorf("failed to insert effect: %w", err)
		}
		return result, nil
	})

	return err
}

// GetEffectByID retrieves an effect by its effectId field
func (r *repository) GetEffectByID(ctx context.Context, effectID string) (*domain.Effect, error) {
	collection := r.getCollection()
	var effect domain.Effect

	result, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		err := collection.FindOne(ctx, bson.M{"effectId": effectID}).Decode(&effect)
		return &effect, err
	})

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("effect not found")
		}
		return nil, fmt.Errorf("failed to get effect: %w", err)
	}

	if result != nil {
		return result.(*domain.Effect), nil
	}

	return &effect, nil
}

// UpdateEffect updates an existing effect
func (r *repository) UpdateEffect(ctx context.Context, effectID string, effect *domain.Effect) error {
	collection := r.getCollection()
	effect.UpdatedAt = time.Now().Unix()

	update := bson.M{"$set": effect}

	_, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		result, err := collection.UpdateOne(
			ctx,
			bson.M{"effectId": effectID},
			update,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update effect: %w", err)
		}
		if result.MatchedCount == 0 {
			return nil, fmt.Errorf("effect not found")
		}
		return result, nil
	})

	return err
}

// DeleteEffect deletes an effect by its effectId
func (r *repository) DeleteEffect(ctx context.Context, effectID string) error {
	collection := r.getCollection()

	_, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		result, err := collection.DeleteOne(ctx, bson.M{"effectId": effectID})
		if err != nil {
			return nil, fmt.Errorf("failed to delete effect: %w", err)
		}
		if result.DeletedCount == 0 {
			return nil, fmt.Errorf("effect not found")
		}
		return result, nil
	})

	return err
}

// ListEffects retrieves effects with filtering and pagination
func (r *repository) ListEffects(ctx context.Context, filter bson.M, page, limit int) ([]domain.Effect, int64, error) {
	collection := r.getCollection()

	skip := int64((page - 1) * limit)

	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "order", Value: 1}, {Key: "created_at", Value: -1}})

	result, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		cursor, err := collection.Find(ctx, filter, findOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to find effects: %w", err)
		}
		defer cursor.Close(ctx)

		var effects []domain.Effect
		if err := cursor.All(ctx, &effects); err != nil {
			return nil, fmt.Errorf("failed to decode effects: %w", err)
		}

		return effects, nil
	})

	if err != nil {
		return nil, 0, err
	}

	effects := result.([]domain.Effect)

	countResult, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		count, err := collection.CountDocuments(ctx, filter)
		return count, err
	})

	if err != nil {
		return nil, 0, fmt.Errorf("failed to count effects: %w", err)
	}

	totalCount := countResult.(int64)

	return effects, totalCount, nil
}

// EffectExists checks if an effect with the given effectId exists
func (r *repository) EffectExists(ctx context.Context, effectID string) (bool, error) {
	collection := r.getCollection()

	result, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		count, err := collection.CountDocuments(ctx, bson.M{"effectId": effectID})
		return count, err
	})

	if err != nil {
		return false, fmt.Errorf("failed to check effect existence: %w", err)
	}

	count := result.(int64)
	return count > 0, nil
}
