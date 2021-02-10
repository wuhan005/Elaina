package db

import (
	"github.com/jackc/pgtype"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type tpl struct {
	gorm.Model

	Name     string
	UID      string
	Language pgtype.TextArray `gorm:"type:text[]"`

	// Limit
	Timeout           int
	MaxCPUs           int
	MaxMemory         int64
	InternetAccess    bool
	DNS               datatypes.JSON `gorm:"type:jsonb"`
	MaxContainer      int
	MaxContainerPerIP int
}
