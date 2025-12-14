package domain

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DurationPolicy defines how long an effect remains active
type DurationPolicy int

const (
	DurationPolicyInstant    DurationPolicy = 0 // Applied once, no duration
	DurationPolicyDuration   DurationPolicy = 1 // Active for a fixed time
	DurationPolicyPersistent DurationPolicy = 2 // Active until explicitly removed
)

// UnmarshalBSONValue implements bson.ValueUnmarshaler to support both string and int
func (d *DurationPolicy) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	raw := bson.RawValue{Type: t, Value: data}

	switch t {
	case bsontype.String:
		var str string
		if err := raw.Unmarshal(&str); err != nil {
			return err
		}
		switch str {
		case "Instant":
			*d = DurationPolicyInstant
		case "Duration":
			*d = DurationPolicyDuration
		case "Persistent":
			*d = DurationPolicyPersistent
		default:
			return fmt.Errorf("invalid DurationPolicy: %s", str)
		}
		return nil
	case bsontype.Int32:
		var i int32
		if err := raw.Unmarshal(&i); err != nil {
			return err
		}
		if i < 0 || i > 2 {
			return fmt.Errorf("invalid DurationPolicy value: %d", i)
		}
		*d = DurationPolicy(i)
		return nil
	case bsontype.Int64:
		var i int64
		if err := raw.Unmarshal(&i); err != nil {
			return err
		}
		if i < 0 || i > 2 {
			return fmt.Errorf("invalid DurationPolicy value: %d", i)
		}
		*d = DurationPolicy(i)
		return nil
	case bsontype.Double:
		var f float64
		if err := raw.Unmarshal(&f); err != nil {
			return err
		}
		*d = DurationPolicy(int(f))
		return nil
	default:
		return fmt.Errorf("cannot decode %v into DurationPolicy", t)
	}
}

// UnmarshalJSON implements json.Unmarshaler to support both string and int
func (d *DurationPolicy) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		switch str {
		case "Instant":
			*d = DurationPolicyInstant
		case "Duration":
			*d = DurationPolicyDuration
		case "Persistent":
			*d = DurationPolicyPersistent
		default:
			return fmt.Errorf("invalid DurationPolicy: %s", str)
		}
		return nil
	}

	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		if i < 0 || i > 2 {
			return fmt.Errorf("invalid DurationPolicy value: %d", i)
		}
		*d = DurationPolicy(i)
		return nil
	}

	return fmt.Errorf("DurationPolicy must be string or int")
}

// EffectCategory defines the type of gameplay effect
type EffectCategory string

const (
	EffectCategoryBuff           EffectCategory = "Buff"
	EffectCategoryDebuff         EffectCategory = "Debuff"
	EffectCategoryDamageOverTime EffectCategory = "DamageOverTime"
	EffectCategoryHealOverTime   EffectCategory = "HealOverTime"
	EffectCategoryShield         EffectCategory = "Shield"
	EffectCategoryStatusEffect   EffectCategory = "StatusEffect"
	EffectCategoryNegativeEffect EffectCategory = "NegativeEffect"
	EffectCategoryPassive        EffectCategory = "Passive"
)

// StatModifier represents a single stat modification
type StatModifier struct {
	Name    string  `bson:"name" json:"name" binding:"required"`
	Flat    float64 `bson:"flat" json:"flat"`
	Percent float64 `bson:"percent" json:"percent"`
}

// EffectMetadata stores additional configuration for effects
type EffectMetadata struct {
	VFX         string   `bson:"vfx,omitempty" json:"vfx,omitempty"`
	SFX         string   `bson:"sfx,omitempty" json:"sfx,omitempty"`
	TagsToApply []string `bson:"tagsToApply,omitempty" json:"tagsToApply,omitempty"`
}

// Effect represents a reusable gameplay effect configuration
type Effect struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	EffectID       string             `bson:"effectId" json:"effectId" binding:"required"`
	Category       EffectCategory     `bson:"category" json:"category" binding:"required"`
	Name           string             `bson:"name" json:"name" binding:"required"`
	Order          int                `bson:"order" json:"order"`
	EffectType     string             `bson:"effectType" json:"effectType"`
	DurationPolicy DurationPolicy     `bson:"durationPolicy" json:"durationPolicy" binding:"required"`
	Duration       float64            `bson:"duration" json:"duration"`
	Interval       float64            `bson:"interval" json:"interval"`
	Stacks         int                `bson:"stacks" json:"stacks"`
	Tags           []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	Modifiers      []StatModifier     `bson:"modifiers,omitempty" json:"modifiers,omitempty"`
	Metadata       EffectMetadata     `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt      int64              `bson:"created_at" json:"created_at"`
	UpdatedAt      int64              `bson:"updated_at" json:"updated_at"`
	CreatedBy      string             `bson:"created_by" json:"created_by"`
	UpdatedBy      string             `bson:"updated_by" json:"updated_by"`
}

// CreateEffectRequest represents a request to create a new effect
type CreateEffectRequest struct {
	EffectID       string         `json:"effectId" binding:"required"`
	Category       EffectCategory `json:"category" binding:"required"`
	Name           string         `json:"name" binding:"required"`
	Order          int            `json:"order"`
	EffectType     string         `json:"effectType"`
	DurationPolicy DurationPolicy `json:"durationPolicy" binding:"required"`
	Duration       float64        `json:"duration"`
	Interval       float64        `json:"interval"`
	Stacks         int            `json:"stacks"`
	Tags           []string       `json:"tags,omitempty"`
	Modifiers      []StatModifier `json:"modifiers,omitempty"`
	Metadata       EffectMetadata `json:"metadata,omitempty"`
}

// UpdateEffectRequest represents a request to update an effect
type UpdateEffectRequest struct {
	Category       *EffectCategory `json:"category,omitempty"`
	Name           *string         `json:"name,omitempty"`
	Order          *int            `json:"order,omitempty"`
	EffectType     *string         `json:"effectType,omitempty"`
	DurationPolicy *DurationPolicy `json:"durationPolicy,omitempty"`
	Duration       *float64        `json:"duration,omitempty"`
	Interval       *float64        `json:"interval,omitempty"`
	Stacks         *int            `json:"stacks,omitempty"`
	Tags           []string        `json:"tags,omitempty"`
	Modifiers      []StatModifier  `json:"modifiers,omitempty"`
	Metadata       *EffectMetadata `json:"metadata,omitempty"`
}

// EffectResponse represents an effect response
type EffectResponse struct {
	EffectID       string         `json:"effectId"`
	Category       EffectCategory `json:"category"`
	Name           string         `json:"name"`
	Order          int            `json:"order"`
	EffectType     string         `json:"effectType,omitempty"`
	DurationPolicy DurationPolicy `json:"durationPolicy"`
	Duration       float64        `json:"duration,omitempty"`
	Interval       float64        `json:"interval,omitempty"`
	Stacks         int            `json:"stacks,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	Modifiers      []StatModifier `json:"modifiers,omitempty"`
	Metadata       EffectMetadata `json:"metadata,omitempty"`
	CreatedAt      int64          `json:"created_at"`
	UpdatedAt      int64          `json:"updated_at"`
	CreatedBy      string         `json:"created_by"`
	UpdatedBy      string         `json:"updated_by"`
}

// EffectListResponse represents a list of effects with pagination
type EffectListResponse struct {
	Items      []EffectResponse `json:"items"`
	Pagination PaginationInfo   `json:"pagination"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// ListEffectsRequest represents a request to list effects with filters
type ListEffectsRequest struct {
	IDs      []string       `json:"ids,omitempty" form:"ids"`
	Category EffectCategory `json:"category,omitempty" form:"category"`
	Tags     []string       `json:"tags,omitempty" form:"tags"`
	Page     int            `json:"page" form:"page"`
	Limit    int            `json:"limit" form:"limit"`
}
