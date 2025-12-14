package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/errors"
	"game-platform/internal/platform/ghttp"
	"game-platform/internal/platform/security"
	"game-platform/internal/services/ability/application"
	"game-platform/internal/services/ability/domain"
)

// Handler handles HTTP requests for ability effects
type Handler struct {
	logger  *logrus.Logger
	service *application.EffectService
}

// NewHandler creates a new effect handler
func NewHandler(logger *logrus.Logger, service *application.EffectService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// CreateEffect handles effect creation requests
func (h *Handler) CreateEffect(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)

	logger := h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"caller":     caller,
		"handler":    "create_effect",
	})

	var req domain.CreateEffectRequest

	if decodedBody, exists := security.GetDecodedBody(ctx); exists {
		if err := json.Unmarshal(decodedBody, &req); err != nil {
			logger.WithError(err).Error("Failed to parse decoded create effect request")
			ghttp.BadRequest(ctx, errors.MsgInvalidRequestBody, nil)
			return
		}
	} else {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.WithError(err).Error("Failed to parse create effect request")
			ghttp.BadRequest(ctx, errors.MsgInvalidRequestBody, nil)
			return
		}
	}

	response, err := h.service.CreateEffect(ctx, &req, caller)
	if err != nil {
		logger.WithError(err).Error("Service error creating effect")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	ghttp.Created(ctx, response)
}

// ListEffects retrieves all effects with filtering and pagination
func (h *Handler) ListEffects(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)

	logger := h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"caller":     caller,
		"handler":    "list_effects",
	})

	var req domain.ListEffectsRequest

	// Parse query parameters
	if idsParam := ctx.Query("ids"); idsParam != "" {
		req.IDs = strings.Split(idsParam, ",")
	}

	if category := ctx.Query("category"); category != "" {
		req.Category = domain.EffectCategory(category)
	}

	if tagsParam := ctx.Query("tags"); tagsParam != "" {
		req.Tags = strings.Split(tagsParam, ",")
	}

	// Parse pagination
	req.Page = 1
	req.Limit = 50
	if page := ctx.Query("page"); page != "" {
		if p, err := parseIntParam(page); err == nil {
			req.Page = p
		}
	}
	if limit := ctx.Query("limit"); limit != "" {
		if l, err := parseIntParam(limit); err == nil {
			req.Limit = l
		}
	}

	response, err := h.service.ListEffects(ctx, &req)
	if err != nil {
		logger.WithError(err).Error("Service error listing effects")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	ghttp.Success(ctx, response)
}

// GetEffectByID retrieves an effect by ID
func (h *Handler) GetEffectByID(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)
	effectID := ctx.Param("id")

	logger := h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"caller":     caller,
		"handler":    "get_effect",
		"effect_id":  effectID,
	})

	if effectID == "" {
		logger.Error("Missing effect ID")
		ghttp.BadRequest(ctx, errors.MsgBadRequest, gin.H{"error": "effect ID is required"})
		return
	}

	response, err := h.service.GetEffectByID(ctx, effectID)
	if err != nil {
		logger.WithError(err).Error("Service error retrieving effect")
		ghttp.NotFound(ctx, "Effect not found")
		return
	}

	ghttp.Success(ctx, response)
}

// UpdateEffect handles effect update requests
func (h *Handler) UpdateEffect(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)
	effectID := ctx.Param("id")

	logger := h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"caller":     caller,
		"handler":    "update_effect",
		"effect_id":  effectID,
	})

	var req domain.UpdateEffectRequest

	if decodedBody, exists := security.GetDecodedBody(ctx); exists {
		if err := json.Unmarshal(decodedBody, &req); err != nil {
			logger.WithError(err).Error("Failed to parse decoded update effect request")
			ghttp.BadRequest(ctx, errors.MsgInvalidRequestBody, nil)
			return
		}
	} else {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.WithError(err).Error("Failed to parse update effect request")
			ghttp.BadRequest(ctx, errors.MsgInvalidRequestBody, nil)
			return
		}
	}

	response, err := h.service.UpdateEffect(ctx, effectID, &req, caller)
	if err != nil {
		logger.WithError(err).Error("Service error updating effect")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	ghttp.Success(ctx, response)
}

// DeleteEffect handles effect deletion requests
func (h *Handler) DeleteEffect(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)
	effectID := ctx.Param("id")

	logger := h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"caller":     caller,
		"handler":    "delete_effect",
		"effect_id":  effectID,
	})

	if effectID == "" {
		logger.Error("Missing effect ID")
		ghttp.BadRequest(ctx, errors.MsgBadRequest, gin.H{"error": "effect ID is required"})
		return
	}

	err := h.service.DeleteEffect(ctx, effectID)
	if err != nil {
		logger.WithError(err).Error("Service error deleting effect")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	ghttp.Success(ctx, gin.H{"message": "Effect deleted successfully"})
}

// parseIntParam safely parses an integer parameter
func parseIntParam(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
