package application

import (
	"context"

	"github.com/sirupsen/logrus"

	"game-platform/internal/services/masterconfig/domain"
	"game-platform/internal/services/masterconfig/resolver"
)

// CharacterService provides application-level operations for character management
// This service is transport-agnostic and can be used by HTTP, gRPC, or other entry points
type CharacterService struct {
	logger   *logrus.Logger
	resolver *resolver.CharacterResolver
}

// NewCharacterService creates a new character service
func NewCharacterService(logger *logrus.Logger, characterResolver *resolver.CharacterResolver) *CharacterService {
	return &CharacterService{
		logger:   logger,
		resolver: characterResolver,
	}
}

// CreateCharacter creates a new character with the given request data
func (s *CharacterService) CreateCharacter(ctx context.Context, req *domain.CreateCharacterRequest, createdBy string) (*domain.CharacterResponse, error) {
	logger := s.logger.WithFields(logrus.Fields{
		"action":       "create_character",
		"character_id": req.CharacterID,
		"created_by":   createdBy,
	})

	logger.Info("Creating character")

	response, err := s.resolver.CreateCharacter(ctx, req, createdBy)
	if err != nil {
		logger.WithError(err).Error("Failed to create character")
		return nil, err
	}

	logger.Info("Character created successfully")
	return response, nil
}

// GetCharacterByCharacterID retrieves a character by its character_id field
func (s *CharacterService) GetCharacterByCharacterID(ctx context.Context, characterID string) (*domain.CharacterResponse, error) {
	logger := s.logger.WithFields(logrus.Fields{
		"action":       "get_character",
		"character_id": characterID,
	})

	logger.Info("Retrieving character")

	response, err := s.resolver.GetCharacterByCharacterID(ctx, characterID)
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve character")
		return nil, err
	}

	logger.Info("Character retrieved successfully")
	return response, nil
}

// UpdateCharacterByCharacterID updates an existing character by character_id field
func (s *CharacterService) UpdateCharacterByCharacterID(ctx context.Context, characterID string, req *domain.UpdateCharacterRequest, updatedBy string) (*domain.CharacterResponse, error) {
	logger := s.logger.WithFields(logrus.Fields{
		"action":       "update_character",
		"character_id": characterID,
		"updated_by":   updatedBy,
	})

	logger.Info("Updating character")

	response, err := s.resolver.UpdateCharacterByCharacterID(ctx, characterID, req, updatedBy)
	if err != nil {
		logger.WithError(err).Error("Failed to update character")
		return nil, err
	}

	logger.Info("Character updated successfully")
	return response, nil
}

// ListCharacters retrieves characters with filtering and pagination
func (s *CharacterService) ListCharacters(ctx context.Context, req *domain.ListCharactersRequest) (*domain.CharacterListResponse, error) {
	logger := s.logger.WithFields(logrus.Fields{
		"action":    "list_characters",
		"page":      req.Page,
		"page_size": req.PageSize,
	})

	logger.Info("Listing characters")

	response, err := s.resolver.ListCharacters(ctx, req)
	if err != nil {
		logger.WithError(err).Error("Failed to list characters")
		return nil, err
	}

	logger.WithField("total_count", response.TotalCount).Info("Characters listed successfully")
	return response, nil
}
