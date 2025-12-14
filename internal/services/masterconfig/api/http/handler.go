package api

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/errors"
	"game-platform/internal/platform/ghttp"
	"game-platform/internal/platform/security"
	"game-platform/internal/services/masterconfig/application"
	"game-platform/internal/services/masterconfig/domain"
)

// Handler handles HTTP requests related to characters
// It is responsible for HTTP-specific concerns: parsing requests, extracting parameters,
// and formatting responses. Business logic is delegated to the service layer.
type Handler struct {
	logger  *logrus.Logger
	service *application.CharacterService
}

// NewHandler creates a new characters handler
func NewHandler(logger *logrus.Logger, service *application.CharacterService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// CreateCharacter handles character creation requests
func (h *Handler) CreateCharacter(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)

	logger := h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"caller":     caller,
		"handler":    "create_character",
	})

	var req domain.CreateCharacterRequest

	// Check if request was decoded by middleware
	if decodedBody, exists := security.GetDecodedBody(ctx); exists {
		if err := json.Unmarshal(decodedBody, &req); err != nil {
			logger.WithError(err).Error("Failed to parse decoded create character request")
			ghttp.BadRequest(ctx, errors.MsgInvalidRequestBody, nil)
			return
		}
	} else {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.WithError(err).Error("Failed to parse create character request")
			ghttp.BadRequest(ctx, errors.MsgInvalidRequestBody, nil)
			return
		}
	}

	// Validate required fields at HTTP layer
	if req.CharacterID == "" {
		logger.Error("Missing required field: CharacterID")
		ghttp.BadRequest(ctx, errors.MsgBadRequest, gin.H{"error": "character_id is required"})
		return
	}

	// Delegate to service
	response, err := h.service.CreateCharacter(ctx, &req, caller)
	if err != nil {
		logger.WithError(err).Error("Service error creating character")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	ghttp.Created(ctx, response)
}

// ListCharacters retrieves all characters with filtering and pagination
func (h *Handler) ListCharacters(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)

	logger := h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"caller":     caller,
		"handler":    "list_characters",
	})

	var req domain.ListCharactersRequest

	// Parse query parameters
	req.Rarity = ctx.Query("rarity")
	req.Element = ctx.Query("element")
	req.Gender = ctx.Query("gender")

	// Parse array parameters
	if roles := ctx.QueryArray("roles"); len(roles) > 0 {
		req.Roles = roles
	}
	if factions := ctx.QueryArray("factions"); len(factions) > 0 {
		req.Factions = factions
	}
	if tags := ctx.QueryArray("tags"); len(tags) > 0 {
		req.Tags = tags
	}

	// Parse pagination
	page := 1
	pageSize := 20
	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}
	if ps := ctx.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil {
			pageSize = parsed
		}
	}
	req.Page = page
	req.PageSize = pageSize

	// Delegate to service
	response, err := h.service.ListCharacters(ctx, &req)
	if err != nil {
		logger.WithError(err).Error("Service error listing characters")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	ghttp.Success(ctx, response)
}

// GetCharacterByCharacterID retrieves a character by Character ID
func (h *Handler) GetCharacterByCharacterID(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)
	characterID := ctx.Param("character_id")

	logger := h.logger.WithFields(logrus.Fields{
		"request_id":   requestID,
		"caller":       caller,
		"handler":      "get_character",
		"character_id": characterID,
	})

	if characterID == "" {
		logger.Error("Missing character_id")
		ghttp.BadRequest(ctx, errors.MsgBadRequest, gin.H{"error": "character_id is required"})
		return
	}

	// Delegate to service
	response, err := h.service.GetCharacterByCharacterID(ctx, characterID)
	if err != nil {
		logger.WithError(err).Error("Service error retrieving character")
		ghttp.NotFound(ctx, "Character not found")
		return
	}

	ghttp.Success(ctx, response)
}

// UpdateCharacterByCharacterID handles character update requests by Character ID
func (h *Handler) UpdateCharacterByCharacterID(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)
	characterID := ctx.Param("character_id")

	logger := h.logger.WithFields(logrus.Fields{
		"request_id":   requestID,
		"caller":       caller,
		"handler":      "update_character",
		"character_id": characterID,
	})

	var req domain.UpdateCharacterRequest

	if decodedBody, exists := security.GetDecodedBody(ctx); exists {
		if err := json.Unmarshal(decodedBody, &req); err != nil {
			logger.WithError(err).Error("Failed to parse decoded update character request")
			ghttp.BadRequest(ctx, errors.MsgInvalidRequestBody, nil)
			return
		}
	} else {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.WithError(err).Error("Failed to parse update character request")
			ghttp.BadRequest(ctx, errors.MsgInvalidRequestBody, nil)
			return
		}
	}

	// Delegate to service
	response, err := h.service.UpdateCharacterByCharacterID(ctx, characterID, &req, caller)
	if err != nil {
		logger.WithError(err).Error("Service error updating character")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	ghttp.Success(ctx, response)
}
