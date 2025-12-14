package application

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/errors"
	"game-platform/internal/platform/ghttp"
	"game-platform/internal/platform/security"
	"game-platform/internal/services/masterconfig/domain"
)

// Controller handles HTTP-level logic for character operations
type Controller struct {
	logger  *logrus.Logger
	service Service
}

// NewController creates a new character controller
func NewController(logger *logrus.Logger, service Service) *Controller {
	return &Controller{
		logger:  logger,
		service: service,
	}
}

// CreateCharacter handles character creation requests
func (c *Controller) CreateCharacter(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)

	logger := c.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"caller":     caller,
		"action":     "create_character",
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

	// Validate required fields
	if req.CharacterID == "" {
		logger.Error("Missing required field: CharacterID")
		ghttp.BadRequest(ctx, errors.MsgBadRequest, gin.H{"error": "character_id is required"})
		return
	}

	// Call service
	response, err := c.service.CreateCharacter(ctx, &req, caller)
	if err != nil {
		logger.WithError(err).Error("Failed to create character")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	logger.Info("Character created successfully")
	ghttp.Created(ctx, response)
}

// GetCharacterByCharacterID retrieves a character by its character_id field
func (c *Controller) GetCharacterByCharacterID(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)
	characterID := ctx.Param("character_id")

	logger := c.logger.WithFields(logrus.Fields{
		"request_id":   requestID,
		"caller":       caller,
		"action":       "get_character_by_character_id",
		"character_id": characterID,
	})

	if characterID == "" {
		logger.Error("Missing character_id")
		ghttp.BadRequest(ctx, errors.MsgBadRequest, gin.H{"error": "character_id is required"})
		return
	}

	response, err := c.service.GetCharacterByCharacterID(ctx, characterID)
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve character")
		ghttp.NotFound(ctx, "Character not found")
		return
	}

	logger.Info("Character retrieved successfully")
	ghttp.Success(ctx, response)
}

// UpdateCharacterByCharacterID handles character update requests by character_id
func (c *Controller) UpdateCharacterByCharacterID(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)
	characterID := ctx.Param("character_id")

	logger := c.logger.WithFields(logrus.Fields{
		"request_id":   requestID,
		"caller":       caller,
		"action":       "update_character_by_character_id",
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

	response, err := c.service.UpdateCharacterByCharacterID(ctx, characterID, &req, caller)
	if err != nil {
		logger.WithError(err).Error("Failed to update character")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	logger.Info("Character updated successfully")
	ghttp.Success(ctx, response)
}

// ListCharacters retrieves all characters with filtering and pagination
func (c *Controller) ListCharacters(ctx *gin.Context) {
	requestID := security.GetRequestID(ctx)
	caller := security.GetCaller(ctx)

	logger := c.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"caller":     caller,
		"action":     "list_characters",
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

	response, err := c.service.ListCharacters(ctx, &req)
	if err != nil {
		logger.WithError(err).Error("Failed to list characters")
		ghttp.InternalServerError(ctx, err.Error())
		return
	}

	logger.Info("Characters listed successfully")
	ghttp.Success(ctx, response)
}
