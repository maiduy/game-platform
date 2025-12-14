package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

// Character represents a playable character with complete stats and configuration
type Character struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CharacterID string             `bson:"characterId" json:"character_id" binding:"required"`
	Meta        Meta               `bson:"meta" json:"meta" binding:"required"`
	BaseStats   BaseStats          `bson:"baseStats" json:"base_stats" binding:"required"`
	Growths     Growths            `bson:"growths" json:"growths" binding:"required"`
	Assets      Assets             `bson:"assets" json:"assets" binding:"required"`
	SkillSet    []string           `bson:"skillSet" json:"skill_set" binding:"required,min=1"`
	CreatedAt   int64              `bson:"created_at" json:"created_at"`
	UpdatedAt   int64              `bson:"updated_at" json:"updated_at"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
	UpdatedBy   string             `bson:"updated_by" json:"updated_by"`
}

// Meta contains display and categorization metadata for the character
type Meta struct {
	NameKey        string   `bson:"nameKey" json:"nameKey" binding:"required"`
	DescriptionKey string   `bson:"descriptionKey" json:"descriptionKey" binding:"required"`
	Icon           string   `bson:"icon" json:"icon" binding:"required"`
	Model          string   `bson:"model" json:"model" binding:"required"`
	Rarity         string   `bson:"rarity" json:"rarity" binding:"required,oneof=R SR SSR"`
	Element        string   `bson:"element" json:"element" binding:"required,oneof=Fire Ice Light Dark Nature Lightning Water Earth"`
	Roles          []string `bson:"roles" json:"roles" binding:"required,min=1"`
	Factions       []string `bson:"factions" json:"factions"`
	Gender         string   `bson:"gender" json:"gender" binding:"required,oneof=Male Female Other"`
	Tags           []string `bson:"tags" json:"tags"`
}

// BaseStats defines character's starting statistics at level 1
type BaseStats struct {
	Level     int     `bson:"level" json:"level" binding:"required"`
	MaxLevel  int     `bson:"maxLevel" json:"maxLevel" binding:"required"`
	HP        int     `bson:"hp" json:"hp" binding:"required,min=1"`
	ATK       int     `bson:"atk" json:"atk" binding:"required,min=1"`
	DEF       int     `bson:"def" json:"def" binding:"required,min=1"`
	SPD       int     `bson:"spd" json:"spd" binding:"required,min=1"`
	Crit      float64 `bson:"crit" json:"crit" binding:"required,min=0,max=1"`
	CritDmg   float64 `bson:"critDmg" json:"critDmg" binding:"required,min=1"`
	Accuracy  float64 `bson:"accuracy" json:"accuracy" binding:"required,min=0,max=1"`
	Evasion   float64 `bson:"evasion" json:"evasion" binding:"required,min=0,max=1"`
	EffectRes float64 `bson:"effectRes" json:"effectRes" binding:"required,min=0,max=1"`
}

// Growths defines character growth curve references
type Growths struct {
	Base      string `bson:"base" json:"base" binding:"required"`
	Assension string `bson:"assension" json:"assension" binding:"required"`
}

// Assets references to character's visual and audio assets
type Assets struct {
	Prefab     string `bson:"prefab" json:"prefab" binding:"required"`
	Animations string `bson:"animations" json:"animations,omitempty"`
	Portrait   string `bson:"portrait" json:"portrait,omitempty"`
	SkillVFX   string `bson:"skill_vfx" json:"skill_vfx,omitempty"`
}

// CreateCharacterRequest represents a request to create a new character
type CreateCharacterRequest struct {
	CharacterID string    `json:"character_id" binding:"required"`
	Meta        Meta      `json:"meta" binding:"required"`
	BaseStats   BaseStats `json:"base_stats" binding:"required"`
	Growths     Growths   `json:"growths" binding:"required"`
	Assets      Assets    `json:"assets" binding:"required"`
	SkillSet    []string  `json:"skill_set" binding:"required,min=1"`
}

// UpdateCharacterRequest represents a request to update a character
type UpdateCharacterRequest struct {
	Meta      *Meta      `json:"meta,omitempty"`
	BaseStats *BaseStats `json:"base_stats,omitempty"`
	Growths   *Growths   `json:"growths,omitempty"`
	Assets    *Assets    `json:"assets,omitempty"`
	SkillSet  []string   `json:"skill_set,omitempty"`
}

// CharacterResponse represents a character response
type CharacterResponse struct {
	//ID          string    `json:"id"`
	CharacterID string    `json:"character_id"`
	Meta        Meta      `json:"meta"`
	BaseStats   BaseStats `json:"baseStats"`
	Growths     Growths   `json:"growths"`
	Assets      Assets    `json:"assets"`
	SkillSet    []string  `json:"skillSet"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`
	UpdatedBy   string    `json:"updated_by"`
}

// CharacterListResponse represents a list of characters with pagination
type CharacterListResponse struct {
	Characters []CharacterResponse `json:"characters"`
	TotalCount int64               `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

// ListCharactersRequest represents a request to list characters with filters, sorting, and pagination
type ListCharactersRequest struct {
	// Filters
	Rarity   string   `json:"rarity,omitempty" form:"rarity"`
	Element  string   `json:"element,omitempty" form:"element"`
	Roles    []string `json:"roles,omitempty" form:"roles"`
	Factions []string `json:"factions,omitempty" form:"factions"`
	Gender   string   `json:"gender,omitempty" form:"gender"`
	Tags     []string `json:"tags,omitempty" form:"tags"`

	// Sorting
	SortBy    string `json:"sort_by,omitempty" form:"sort_by"`       // Field to sort by (e.g., "created_at", "rarity", "name_key")
	SortOrder string `json:"sort_order,omitempty" form:"sort_order"` // "asc" or "desc" (default: "desc")

	// Pagination
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}
