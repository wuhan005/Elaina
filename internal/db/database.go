package db

import (
	"os"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() (*gorm.DB, error) {
	dsn := os.ExpandEnv("postgres://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE?sslmode=$PGSSLMODE")
	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)
	if err != nil {
		return nil, errors.Wrap(err, "open database")
	}

	err = db.AutoMigrate(&Tpl{}, &Sandbox{})
	if err != nil {
		return nil, errors.Wrap(err, "auto migrate")
	}

	Tpls = &tpls{DB: db}
	Sandboxes = &sandboxes{DB: db}

	return db, nil
}
