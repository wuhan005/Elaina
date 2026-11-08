package db

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/dbutil"
)

func Init() (*gorm.DB, error) {
	dsn := config.Postgres.DSN

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		// Sandbox saves include associations and must remain atomic.
		SkipDefaultTransaction: false,
		NowFunc:                dbutil.Now,
		Logger: logger.New(
			logrus.New(),
			logger.Config{
				SlowThreshold:             3 * time.Second,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		return nil, errors.Wrap(err, "open database")
	}

	if err := db.AutoMigrate(
		&Tpl{},
		&Sandbox{},
	); err != nil {
		return nil, errors.Wrap(err, "auto migrate")
	}

	Tpls = NewTplStore(db)
	Sandboxes = NewSandboxStore(db)

	return db, nil
}
