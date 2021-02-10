package db

import (
	"gorm.io/gorm"
)

type Sandbox struct {
	gorm.Model

	UID         string `gorm:"NOT NULL" json:"uid"`
	Name        string `json:"name"`
	TemplateID  uint   `gorm:"NOT NULL" json:"template_id"`
	Template    *Tpl   `gorm:"ForeignKey:TemplateID" json:"template"`
	Placeholder string `json:"placeholder"`
	Editable    bool   `json:"editable"`
}
