package db

import (
	"gorm.io/gorm"
)

type sandbox struct {
	gorm.Model

	UID         string `gorm:"NOT NULL"`
	TemplateID  uint   `gorm:"NOT NULL"`
	Template    *tpl   `gorm:"ForeignKey:TemplateID"`
	Placeholder string
	Editable    bool
}


