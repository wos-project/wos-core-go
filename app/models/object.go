package models

import (
	"time"

	"gorm.io/gorm"
)

type Object struct {
	gorm.Model
	Cid            string    `gorm:"column:cid; index" binding:"required"`
	OwnerUid       string    `gorm:"column:owner_uid; index" binding:"required"`
	OwnerProvider  string    `gorm:"column:owner_provider; index" binding:"required"`
	Name           string    `gorm:"column:name; index" binding:"required"`
	Description    string    `gorm:"description; index"`
	CoverImageUri  string    `gorm:"cover_image_uri"`
	CreatedAtInner time.Time `gorm:"column:created_at_inner"`
	Body           JSONMap   `gorm:"column:body"`
	Files          JSONMap   `gorm:"column:files"`
}
