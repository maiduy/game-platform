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
	"game-platform/internal/services/masterconfig/domain"
)

const (
	collectionName = "characters"
)

// Repository defines the interface for character data access
type Repository interface {
	CreateCharacter(ctx context.Context, character *domain.Character) error
	GetCharacterByCharacterID(ctx context.Context, characterID string) (*domain.Character, error)
	UpdateCharacter(ctx context.Context, id string, character *domain.Character) error
	ListCharacters(ctx context.Context, filter bson.M, page, pageSize int) ([]domain.Character, int64, error)
	CharacterExists(ctx context.Context, characterID string) (bool, error)
}

// repository implements Repository interface using MongoDB
type repository struct {
	mongoConn *pool.MongoDBConnection
}

// NewRepository creates a new character repository
func NewRepository(mongoConn *pool.MongoDBConnection) Repository {
	return &repository{
		mongoConn: mongoConn,
	}
}

// getCollection returns the characters collection
func (r *repository) getCollection() *mongo.Collection {
	return r.mongoConn.GetDefaultCollection(collectionName)
}

// CreateCharacter creates a new character in the database
func (r *repository) CreateCharacter(ctx context.Context, character *domain.Character) error {
	collection := r.getCollection()

	// Generate new ObjectID if not set
	if character.ID.IsZero() {
		character.ID = primitive.NewObjectID()
	}

	// Execute with retry logic and circuit breaker
	_, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		result, err := collection.InsertOne(ctx, character)
		if err != nil {
			return nil, fmt.Errorf("failed to insert character: %w", err)
		}
		return result, nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetCharacterByCharacterID retrieves a character by its character_id field
func (r *repository) GetCharacterByCharacterID(ctx context.Context, characterID string) (*domain.Character, error) {
	collection := r.getCollection()
	var character domain.Character

	result, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		err := collection.FindOne(ctx, bson.M{"characterId": characterID}).Decode(&character)
		return &character, err
	})

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("character not found")
		}
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	if result != nil {
		return result.(*domain.Character), nil
	}

	return &character, nil
}

// UpdateCharacter updates an existing character
func (r *repository) UpdateCharacter(ctx context.Context, id string, character *domain.Character) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid character ID: %w", err)
	}

	collection := r.getCollection()
	character.UpdatedAt = time.Now().Unix()

	// Create update document
	update := bson.M{
		"$set": character,
	}

	_, err = r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		result, err := collection.UpdateOne(
			ctx,
			bson.M{"_id": objectID},
			update,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update character: %w", err)
		}
		if result.MatchedCount == 0 {
			return nil, fmt.Errorf("character not found")
		}
		return result, nil
	})

	return err
}

// ListCharacters retrieves characters with filtering and pagination
func (r *repository) ListCharacters(ctx context.Context, filter bson.M, page, pageSize int) ([]domain.Character, int64, error) {
	collection := r.getCollection()

	// Calculate skip
	skip := int64((page - 1) * pageSize)

	// Set up find options
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(int64(pageSize))
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	// Execute with retry
	result, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		cursor, err := collection.Find(ctx, filter, findOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to find characters: %w", err)
		}
		defer cursor.Close(ctx)

		var characters []domain.Character
		if err := cursor.All(ctx, &characters); err != nil {
			return nil, fmt.Errorf("failed to decode characters: %w", err)
		}

		return characters, nil
	})

	if err != nil {
		return nil, 0, err
	}

	characters := result.([]domain.Character)

	// Get total count
	countResult, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		count, err := collection.CountDocuments(ctx, filter)
		return count, err
	})

	if err != nil {
		return nil, 0, fmt.Errorf("failed to count characters: %w", err)
	}

	totalCount := countResult.(int64)

	return characters, totalCount, nil
}

// CharacterExists checks if a character with the given character_id exists
func (r *repository) CharacterExists(ctx context.Context, characterID string) (bool, error) {
	collection := r.getCollection()

	result, err := r.mongoConn.ExecuteWithRetry(func() (interface{}, error) {
		count, err := collection.CountDocuments(ctx, bson.M{"characterId": characterID})
		return count, err
	})

	if err != nil {
		return false, fmt.Errorf("failed to check character existence: %w", err)
	}

	count := result.(int64)
	return count > 0, nil
}
