package models

import (
	"gorm.io/gorm"
)

const (
	MediaUploadEnabled  = 1 // enabled
	MediaUploadDisabled = 2 // deleted
)

type MediaUpload struct {
	gorm.Model
	SessionID string  `gorm:"column:session_id; index"`
	Path      string  `gorm:"column:path; index"`
	Status    byte    `gorm:"column:status`
	Metadata  JSONMap `gorm:"column:metadata"`
}
