package resolver

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"game-platform/internal/services/ability/domain"
	"game-platform/internal/services/ability/repository"
)

var effectIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// EffectResolver handles effect business logic operations
type EffectResolver struct {
	repository repository.Repository
}

// NewEffectResolver creates a new effect resolver
func NewEffectResolver(repo repository.Repository) *EffectResolver {
	return &EffectResolver{
		repository: repo,
	}
}

// CreateEffect creates a new effect with validation
func (r *EffectResolver) CreateEffect(ctx context.Context, req *domain.CreateEffectRequest, createdBy string) (*domain.EffectResponse, error) {
	// Validate required fields
	if err := r.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check if effect already exists
	exists, err := r.repository.EffectExists(ctx, req.EffectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check effect existence: %w", err)
	}
	if exists {
		return nil, errors.New("effect with this effectId already exists")
	}

	now := time.Now().Unix()
	effect := &domain.Effect{
		EffectID:       req.EffectID,
		Category:       req.Category,
		Name:           req.Name,
		Order:          req.Order,
		EffectType:     req.EffectType,
		DurationPolicy: req.DurationPolicy,
		Duration:       req.Duration,
		Interval:       req.Interval,
		Stacks:         req.Stacks,
		Tags:           req.Tags,
		Modifiers:      req.Modifiers,
		Metadata:       req.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedBy:      createdBy,
		UpdatedBy:      createdBy,
	}

	if err := r.repository.CreateEffect(ctx, effect); err != nil {
		return nil, fmt.Errorf("failed to create effect: %w", err)
	}

	return r.effectToResponse(effect), nil
}

// GetEffectByID retrieves an effect by its effectId
func (r *EffectResolver) GetEffectByID(ctx context.Context, effectID string) (*domain.EffectResponse, error) {
	if effectID == "" {
		return nil, errors.New("effectId is required")
	}

	effect, err := r.repository.GetEffectByID(ctx, effectID)
	if err != nil {
		return nil, err
	}

	return r.effectToResponse(effect), nil
}

// UpdateEffect updates an existing effect
func (r *EffectResolver) UpdateEffect(ctx context.Context, effectID string, req *domain.UpdateEffectRequest, updatedBy string) (*domain.EffectResponse, error) {
	if effectID == "" {
		return nil, errors.New("effectId is required")
	}

	// Get existing effect
	existing, err := r.repository.GetEffectByID(ctx, effectID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Category != nil {
		if !r.isValidCategory(*req.Category) {
			return nil, errors.New("invalid category")
		}
		existing.Category = *req.Category
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Order != nil {
		if *req.Order < 0 || *req.Order > 100 {
			return nil, errors.New("order must be between 0 and 100")
		}
		existing.Order = *req.Order
	}
	if req.EffectType != nil {
		existing.EffectType = *req.EffectType
	}
	if req.DurationPolicy != nil {
		if !r.isValidDurationPolicy(*req.DurationPolicy) {
			return nil, errors.New("invalid durationPolicy")
		}
		existing.DurationPolicy = *req.DurationPolicy
	}
	if req.Duration != nil {
		if *req.Duration < 0 {
			return nil, errors.New("duration must be >= 0")
		}
		existing.Duration = *req.Duration
	}
	if req.Interval != nil {
		if *req.Interval < 0 {
			return nil, errors.New("interval must be >= 0")
		}
		existing.Interval = *req.Interval
	}
	if req.Stacks != nil {
		if *req.Stacks < 0 {
			return nil, errors.New("stacks must be >= 0")
		}
		existing.Stacks = *req.Stacks
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.Modifiers != nil {
		existing.Modifiers = req.Modifiers
	}
	if req.Metadata != nil {
		existing.Metadata = *req.Metadata
	}

	existing.UpdatedBy = updatedBy
	existing.UpdatedAt = time.Now().Unix()

	if err := r.repository.UpdateEffect(ctx, effectID, existing); err != nil {
		return nil, fmt.Errorf("failed to update effect: %w", err)
	}

	return r.effectToResponse(existing), nil
}

// DeleteEffect deletes an effect
func (r *EffectResolver) DeleteEffect(ctx context.Context, effectID string) error {
	if effectID == "" {
		return errors.New("effectId is required")
	}

	// Check if effect exists
	exists, err := r.repository.EffectExists(ctx, effectID)
	if err != nil {
		return fmt.Errorf("failed to check effect existence: %w", err)
	}
	if !exists {
		return errors.New("effect not found")
	}

	return r.repository.DeleteEffect(ctx, effectID)
}

// ListEffects retrieves effects with filtering and pagination
func (r *EffectResolver) ListEffects(ctx context.Context, req *domain.ListEffectsRequest) (*domain.EffectListResponse, error) {
	filter := bson.M{}

	if len(req.IDs) > 0 {
		filter["effectId"] = bson.M{"$in": req.IDs}
	}

	if req.Category != "" {
		filter["category"] = req.Category
	}

	if len(req.Tags) > 0 {
		filter["tags"] = bson.M{"$in": req.Tags}
	}

	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	effects, totalCount, err := r.repository.ListEffects(ctx, filter, req.Page, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list effects: %w", err)
	}

	effectResponses := make([]domain.EffectResponse, len(effects))
	for i, effect := range effects {
		effectResponses[i] = *r.effectToResponse(&effect)
	}

	totalPages := int(totalCount) / req.Limit
	if int(totalCount)%req.Limit > 0 {
		totalPages++
	}

	return &domain.EffectListResponse{
		Items: effectResponses,
		Pagination: domain.PaginationInfo{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// validateCreateRequest validates the create effect request
func (r *EffectResolver) validateCreateRequest(req *domain.CreateEffectRequest) error {
	if req.EffectID == "" {
		return errors.New("effectId is required")
	}

	if len(req.EffectID) > 50 {
		return errors.New("effectId must be max 50 characters")
	}

	if !effectIDPattern.MatchString(req.EffectID) {
		return errors.New("effectId must match pattern ^[a-zA-Z0-9_]+$")
	}

	if req.Name == "" {
		return errors.New("name is required")
	}

	if !r.isValidCategory(req.Category) {
		return errors.New("category is required and must be valid")
	}

	if !r.isValidDurationPolicy(req.DurationPolicy) {
		return errors.New("durationPolicy is required and must be valid")
	}

	if req.Duration < 0 {
		return errors.New("duration must be >= 0")
	}

	if req.Interval < 0 {
		return errors.New("interval must be >= 0")
	}

	if req.Stacks < 0 {
		return errors.New("stacks must be >= 0")
	}

	if req.Order < 0 || req.Order > 100 {
		return errors.New("order must be between 0 and 100")
	}

	return nil
}

// isValidCategory checks if the category is valid
func (r *EffectResolver) isValidCategory(category domain.EffectCategory) bool {
	validCategories := []domain.EffectCategory{
		domain.EffectCategoryBuff,
		domain.EffectCategoryDebuff,
		domain.EffectCategoryDamageOverTime,
		domain.EffectCategoryHealOverTime,
		domain.EffectCategoryShield,
		domain.EffectCategoryStatusEffect,
		domain.EffectCategoryNegativeEffect,
		domain.EffectCategoryPassive,
	}

	categoryStr := strings.TrimSpace(string(category))
	for _, valid := range validCategories {
		if string(valid) == categoryStr {
			return true
		}
	}
	return false
}

// isValidDurationPolicy checks if the duration policy is valid
func (r *EffectResolver) isValidDurationPolicy(policy domain.DurationPolicy) bool {
	return policy == domain.DurationPolicyInstant ||
		policy == domain.DurationPolicyDuration ||
		policy == domain.DurationPolicyPersistent
}

// effectToResponse converts an Effect to EffectResponse
func (r *EffectResolver) effectToResponse(effect *domain.Effect) *domain.EffectResponse {
	return &domain.EffectResponse{
		EffectID:       effect.EffectID,
		Category:       effect.Category,
		Name:           effect.Name,
		Order:          effect.Order,
		EffectType:     effect.EffectType,
		DurationPolicy: effect.DurationPolicy,
		Duration:       effect.Duration,
		Interval:       effect.Interval,
		Stacks:         effect.Stacks,
		Tags:           effect.Tags,
		Modifiers:      effect.Modifiers,
		Metadata:       effect.Metadata,
		CreatedAt:      effect.CreatedAt,
		UpdatedAt:      effect.UpdatedAt,
		CreatedBy:      effect.CreatedBy,
		UpdatedBy:      effect.UpdatedBy,
	}
}
