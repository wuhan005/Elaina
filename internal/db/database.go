package db

import (
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/config"
)

func Init() (*gorm.DB, error) {
	dsn := config.Postgres.DSN

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)
	if err != nil {
		return nil, errors.Wrap(err, "open database")
	}

	if err := db.AutoMigrate(&Tpl{}, &Sandbox{}); err != nil {
		return nil, errors.Wrap(err, "auto migrate")
	}

	Tpls = NewTplStore(db)
	Sandboxes = NewSandboxStore(db)

	return db, nil
}
