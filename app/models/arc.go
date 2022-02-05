package models

import (
	"gorm.io/gorm"
)

type Arc struct {
	gorm.Model
	Object
}
