package db

import (
	"gorm.io/gorm"
)

type sandbox struct {
	gorm.Model

	UID         string `gorm:"NOT NULL" json:"uid"`
	TemplateID  uint   `gorm:"NOT NULL" json:"template_id"`
	Template    *tpl   `gorm:"ForeignKey:TemplateID" json:"template"`
	Placeholder string `json:"placeholder"`
	Editable    bool   `json:"editable"`
}
