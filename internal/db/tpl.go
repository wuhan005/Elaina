package db

import (
	"github.com/jackc/pgtype"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Tpl struct {
	gorm.Model

	Name     string           `json:"name"`
	Language pgtype.TextArray `gorm:"type:text[]" json:"language"`

	// Limit
	Timeout           int            `json:"timeout"`
	MaxCPUs           int64          `json:"max_cpus"`
	MaxMemory         int64          `json:"max_memory"`
	InternetAccess    bool           `json:"internet_access"`
	DNS               datatypes.JSON `gorm:"type:jsonb" json:"dns"`
	MaxContainer      int64          `json:"max_container"`
	MaxContainerPerIP int64          `json:"max_container_per_ip"`
}
