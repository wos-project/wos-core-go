package models

import (
	"gorm.io/gorm"
)

type Pin struct {
	gorm.Model
	Object
	Location PointGeo `gorm:"column:location; index"`
}
