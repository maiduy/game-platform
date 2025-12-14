package api

import (
	"github.com/gin-gonic/gin"

	controller "game-platform/internal/services/masterconfig/application"
)

// Handler handles HTTP requests related to characters
type Handler struct {
	controller *controller.Controller
}

// NewHandler creates a new characters handler
func NewHandler(controller *controller.Controller) *Handler {
	return &Handler{
		controller: controller,
	}
}

// CreateCharacter handles character creation requests
func (h *Handler) CreateCharacter(ctx *gin.Context) {
	h.controller.CreateCharacter(ctx)
}

// ListCharacters retrieves all characters with filtering and pagination
func (h *Handler) ListCharacters(ctx *gin.Context) {
	h.controller.ListCharacters(ctx)
}

// GetCharacterByCharacterID retrieves a character by Character ID
func (h *Handler) GetCharacterByCharacterID(ctx *gin.Context) {
	h.controller.GetCharacterByCharacterID(ctx)
}

// UpdateCharacterByCharacterID handles character update requests by Character ID
func (h *Handler) UpdateCharacterByCharacterID(ctx *gin.Context) {
	h.controller.UpdateCharacterByCharacterID(ctx)
}
