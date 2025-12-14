package application

import (
	"context"

	"game-platform/internal/services/masterconfig/domain"

	"go.mongodb.org/mongo-driver/bson"
)

// Repository defines the interface for character data access
type Repository interface {
	CreateCharacter(ctx context.Context, character *domain.Character) error
	GetCharacterByCharacterID(ctx context.Context, characterID string) (*domain.Character, error)
	UpdateCharacter(ctx context.Context, id string, character *domain.Character) error
	ListCharacters(ctx context.Context, filter bson.M, page, pageSize int) ([]domain.Character, int64, error)
	CharacterExists(ctx context.Context, characterID string) (bool, error)
}
