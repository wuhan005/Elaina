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
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  config.Postgres.DSN,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		NowFunc: dbutil.Now,
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
	if err := db.Exec(`ALTER TABLE tpls
		DROP COLUMN IF EXISTS max_cpu,
		DROP COLUMN IF EXISTS max_cpus,
		DROP COLUMN IF EXISTS max_container,
		DROP COLUMN IF EXISTS max_container_per_ip,
		DROP COLUMN IF EXISTS dns`).Error; err != nil {
		return nil, errors.Wrap(err, "drop retired template columns")
	}

	Tpls = NewTplStore(db)
	Sandboxes = NewSandboxStore(db)

	return db, nil
}
