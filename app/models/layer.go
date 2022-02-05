package models

import (
	"gorm.io/gorm"
)

type Layer struct {
	gorm.Model
	Uid  string `gorm:"column:uid; index"`
	Name string `gorm:"column:name"`
}
