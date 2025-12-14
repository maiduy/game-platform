package application

import (
	"context"

	"github.com/sirupsen/logrus"

	"game-platform/internal/services/ability/domain"
	"game-platform/internal/services/ability/resolver"
)

// EffectService provides application-level operations for effect management
type EffectService struct {
	logger   *logrus.Logger
	resolver *resolver.EffectResolver
}

// NewEffectService creates a new effect service
func NewEffectService(logger *logrus.Logger, effectResolver *resolver.EffectResolver) *EffectService {
	return &EffectService{
		logger:   logger,
		resolver: effectResolver,
	}
}

// CreateEffect creates a new effect
func (s *EffectService) CreateEffect(ctx context.Context, req *domain.CreateEffectRequest, createdBy string) (*domain.EffectResponse, error) {
	logger := s.logger.WithFields(logrus.Fields{
		"action":     "create_effect",
		"effect_id":  req.EffectID,
		"created_by": createdBy,
	})

	logger.Info("Creating effect")

	response, err := s.resolver.CreateEffect(ctx, req, createdBy)
	if err != nil {
		logger.WithError(err).Error("Failed to create effect")
		return nil, err
	}

	logger.Info("Effect created successfully")
	return response, nil
}

// GetEffectByID retrieves an effect by its ID
func (s *EffectService) GetEffectByID(ctx context.Context, effectID string) (*domain.EffectResponse, error) {
	logger := s.logger.WithFields(logrus.Fields{
		"action":    "get_effect",
		"effect_id": effectID,
	})

	logger.Info("Retrieving effect")

	response, err := s.resolver.GetEffectByID(ctx, effectID)
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve effect")
		return nil, err
	}

	logger.Info("Effect retrieved successfully")
	return response, nil
}

// UpdateEffect updates an existing effect
func (s *EffectService) UpdateEffect(ctx context.Context, effectID string, req *domain.UpdateEffectRequest, updatedBy string) (*domain.EffectResponse, error) {
	logger := s.logger.WithFields(logrus.Fields{
		"action":     "update_effect",
		"effect_id":  effectID,
		"updated_by": updatedBy,
	})

	logger.Info("Updating effect")

	response, err := s.resolver.UpdateEffect(ctx, effectID, req, updatedBy)
	if err != nil {
		logger.WithError(err).Error("Failed to update effect")
		return nil, err
	}

	logger.Info("Effect updated successfully")
	return response, nil
}

// DeleteEffect deletes an effect
func (s *EffectService) DeleteEffect(ctx context.Context, effectID string) error {
	logger := s.logger.WithFields(logrus.Fields{
		"action":    "delete_effect",
		"effect_id": effectID,
	})

	logger.Info("Deleting effect")

	err := s.resolver.DeleteEffect(ctx, effectID)
	if err != nil {
		logger.WithError(err).Error("Failed to delete effect")
		return err
	}

	logger.Info("Effect deleted successfully")
	return nil
}

// ListEffects retrieves effects with filtering and pagination
func (s *EffectService) ListEffects(ctx context.Context, req *domain.ListEffectsRequest) (*domain.EffectListResponse, error) {
	logger := s.logger.WithFields(logrus.Fields{
		"action": "list_effects",
		"page":   req.Page,
		"limit":  req.Limit,
	})

	logger.Info("Listing effects")

	response, err := s.resolver.ListEffects(ctx, req)
	if err != nil {
		logger.WithError(err).Error("Failed to list effects")
		return nil, err
	}

	logger.WithField("total", response.Pagination.Total).Info("Effects listed successfully")
	return response, nil
}
