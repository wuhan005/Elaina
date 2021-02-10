package db

import (
	"os"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() error {
	dsn := os.ExpandEnv("postgres://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE?sslmode=$PGSSLMODE")
	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&tpl{})
	if err != nil {
		return errors.Wrap(err, "auto migrate")
	}

	Tpls = &tpls{DB: db}

	return nil
}
