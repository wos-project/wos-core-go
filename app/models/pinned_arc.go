package models

import (
	"gorm.io/gorm"
)

type PinnedArc struct {
	gorm.Model
	Object
	Arc   Arc  `gorm:"foreignkey:ArcId; references: id; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ArcId uint `gorm:"column:arc_id; index"`
	Pin   Pin  `gorm:"foreignkey:PinId; references: id; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PinId uint `gorm:"column:pin_id; index"`
}
