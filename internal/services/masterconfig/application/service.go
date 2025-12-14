package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"game-platform/internal/services/masterconfig/domain"

	"go.mongodb.org/mongo-driver/bson"
)

// Service defines the interface for character business logic
type Service interface {
	// Character CRUD operations
	CreateCharacter(ctx context.Context, req *domain.CreateCharacterRequest, createdBy string) (*domain.CharacterResponse, error)
	GetCharacterByCharacterID(ctx context.Context, characterID string) (*domain.CharacterResponse, error)
	UpdateCharacterByCharacterID(ctx context.Context, characterID string, req *domain.UpdateCharacterRequest, updatedBy string) (*domain.CharacterResponse, error)
	ListCharacters(ctx context.Context, req *domain.ListCharactersRequest) (*domain.CharacterListResponse, error)
}

// service implements Service interface
type service struct {
	repository Repository
}

// NewService creates a new character service
func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

// CreateCharacter creates a new character
func (s *service) CreateCharacter(ctx context.Context, req *domain.CreateCharacterRequest, createdBy string) (*domain.CharacterResponse, error) {
	// Validate required fields
	if req.CharacterID == "" {
		return nil, errors.New("character_id is required")
	}

	// Check if character already exists
	exists, err := s.repository.CharacterExists(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check character existence: %w", err)
	}
	if exists {
		return nil, errors.New("character with this character_id already exists")
	}

	// Validate rarity
	if req.Meta.Rarity != "R" && req.Meta.Rarity != "SR" && req.Meta.Rarity != "SSR" {
		return nil, errors.New("rarity must be R, SR, or SSR")
	}

	// Validate element
	validElements := map[string]bool{
		"Fire": true, "Ice": true, "Light": true, "Dark": true,
		"Nature": true, "Lightning": true, "Water": true, "Earth": true,
	}
	if !validElements[req.Meta.Element] {
		return nil, errors.New("invalid element")
	}

	// Validate gender
	if req.Meta.Gender != "Male" && req.Meta.Gender != "Female" && req.Meta.Gender != "Other" {
		return nil, errors.New("gender must be Male, Female, or Other")
	}

	// Validate skill set
	if len(req.SkillSet) == 0 {
		return nil, errors.New("at least one skill is required in skill_set")
	}

	// Create the character
	now := time.Now().Unix()
	character := &domain.Character{
		CharacterID: req.CharacterID,
		Meta:        req.Meta,
		BaseStats:   req.BaseStats,
		Growths:     req.Growths,
		Assets:      req.Assets,
		SkillSet:    req.SkillSet,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	if err := s.repository.CreateCharacter(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	return s.characterToResponse(character), nil
}

// GetCharacterByCharacterID retrieves a character by its character_id field
func (s *service) GetCharacterByCharacterID(ctx context.Context, characterID string) (*domain.CharacterResponse, error) {
	if characterID == "" {
		return nil, errors.New("character_id is required")
	}

	character, err := s.repository.GetCharacterByCharacterID(ctx, characterID)
	if err != nil {
		return nil, err
	}

	return s.characterToResponse(character), nil
}

// UpdateCharacterByCharacterID updates an existing character by character_id field
func (s *service) UpdateCharacterByCharacterID(ctx context.Context, characterID string, req *domain.UpdateCharacterRequest, updatedBy string) (*domain.CharacterResponse, error) {
	if characterID == "" {
		return nil, errors.New("character_id is required")
	}

	// Get existing character by character_id
	existing, err := s.repository.GetCharacterByCharacterID(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Meta != nil {
		existing.Meta = *req.Meta
	}
	if req.BaseStats != nil {
		existing.BaseStats = *req.BaseStats
	}
	if req.Growths != nil {
		existing.Growths = *req.Growths
	}
	if req.Assets != nil {
		existing.Assets = *req.Assets
	}
	if req.SkillSet != nil && len(req.SkillSet) > 0 {
		existing.SkillSet = req.SkillSet
	}

	existing.UpdatedBy = updatedBy
	existing.UpdatedAt = time.Now().Unix()

	// Use the MongoDB ObjectID to update
	if err := s.repository.UpdateCharacter(ctx, existing.ID.Hex(), existing); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return s.characterToResponse(existing), nil
}

// ListCharacters retrieves characters with filtering and pagination
func (s *service) ListCharacters(ctx context.Context, req *domain.ListCharactersRequest) (*domain.CharacterListResponse, error) {
	// Build filter
	filter := bson.M{}

	if req.Rarity != "" {
		filter["meta.rarity"] = req.Rarity
	}

	if req.Element != "" {
		filter["meta.element"] = req.Element
	}

	if req.Gender != "" {
		filter["meta.gender"] = req.Gender
	}

	if len(req.Roles) > 0 {
		filter["meta.roles"] = bson.M{"$in": req.Roles}
	}

	if len(req.Factions) > 0 {
		filter["meta.factions"] = bson.M{"$in": req.Factions}
	}

	if len(req.Tags) > 0 {
		filter["meta.tags"] = bson.M{"$in": req.Tags}
	}

	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // Max page size
	}

	characters, totalCount, err := s.repository.ListCharacters(ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}

	// Convert to response format
	characterResponses := make([]domain.CharacterResponse, len(characters))
	for i, char := range characters {
		characterResponses[i] = *s.characterToResponse(&char)
	}

	return &domain.CharacterListResponse{
		Characters: characterResponses,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// characterToResponse converts a Character to CharacterResponse
func (s *service) characterToResponse(character *domain.Character) *domain.CharacterResponse {
	// var id string
	// if !character.ID.IsZero() {
	// 	id = character.ID.Hex()
	// } else {
	// 	id = primitive.NewObjectID().Hex()
	// }

	return &domain.CharacterResponse{
		//ID:          id,
		CharacterID: character.CharacterID,
		Meta:        character.Meta,
		BaseStats:   character.BaseStats,
		Growths:     character.Growths,
		Assets:      character.Assets,
		SkillSet:    character.SkillSet,
		CreatedAt:   character.CreatedAt,
		UpdatedAt:   character.UpdatedAt,
		CreatedBy:   character.CreatedBy,
		UpdatedBy:   character.UpdatedBy,
	}
}
