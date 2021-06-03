package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Edition string

const (
	JavaEdition    Edition = "java"
	BedrockEdition Edition = "bedrock"
)

type Game struct {
	gorm.Model
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID
	Name            string  `gorm:"not null;unique;size:255;default:null"`
	CustomSubdomain *string `gorm:"size:63;default:null"`
	Edition         Edition
}
