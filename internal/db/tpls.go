package db

import (
	"github.com/jackc/pgtype"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TplStore interface {
	// GetByID returns a template with the given id.
	GetByID(id uint) (*Tpl, error)
	// ListAll returns all the templates.
	ListAll() ([]*Tpl, error)
	// Create creates a new template with the given options.
	Create(opts CreateTplOptions) error
	// Update edits a new template with the given options.
	Update(opts UpdateTplOptions) error
	// Delete deletes a template with the given id.
	Delete(id uint) error
}

var Tpls TplStore

var _ TplStore = (*tpls)(nil)

type tpls struct {
	*gorm.DB
}

func (db *tpls) GetByID(id uint) (*Tpl, error) {
	var template Tpl
	return &template, db.Model(&Tpl{}).Where("id = ?", id).First(&template).Error
}

func (db *tpls) ListAll() ([]*Tpl, error) {
	var templates []*Tpl
	err := db.Model(&Tpl{}).Find(&templates).Error
	return templates, err
}

type CreateTplOptions struct {
	Name              string
	Language          []string
	Timeout           int
	MaxCPUs           int
	MaxMemory         int64
	InternetAccess    bool
	DNS               map[string]string
	MaxContainer      int
	MaxContainerPerIP int
}

func (db *tpls) Create(opts CreateTplOptions) error {
	languages := pgtype.TextArray{}
	if err := languages.Set(opts.Language); err != nil {
		return errors.Wrap(err, "set language")
	}

	dnsValue, err := jsoniter.Marshal(opts.DNS)
	if err != nil {
		return errors.Wrap(err, "marshal dns")
	}
	dns := datatypes.JSON{}
	if err := dns.Scan(dnsValue); err != nil {
		return errors.Wrap(err, "marshal DNS JSONs")
	}

	return db.DB.Create(&Tpl{
		Name:              opts.Name,
		Language:          languages,
		Timeout:           opts.Timeout,
		MaxCPUs:           opts.MaxCPUs,
		MaxMemory:         opts.MaxMemory,
		InternetAccess:    opts.InternetAccess,
		DNS:               dns,
		MaxContainer:      opts.MaxContainer,
		MaxContainerPerIP: opts.MaxContainerPerIP,
	}).Error
}

type UpdateTplOptions struct {
	ID                uint
	Name              string
	Language          []string
	Timeout           int
	MaxCPUs           int
	MaxMemory         int64
	InternetAccess    bool
	DNS               map[string]string
	MaxContainer      int
	MaxContainerPerIP int
}

func (db *tpls) Update(opts UpdateTplOptions) error {
	_, err := db.GetByID(opts.ID)
	if err != nil {
		return errors.New("template not existed")
	}

	languages := pgtype.TextArray{}
	if err := languages.Set(opts.Language); err != nil {
		return errors.Wrap(err, "set language")
	}

	dnsValue, err := jsoniter.Marshal(opts.DNS)
	if err != nil {
		return errors.Wrap(err, "marshal dns")
	}
	dns := datatypes.JSON{}
	if err := dns.Scan(dnsValue); err != nil {
		return errors.Wrap(err, "marshal DNS JSONs")
	}

	return db.DB.Model(&Tpl{}).Where(&Tpl{
		Model: gorm.Model{
			ID: opts.ID,
		},
	}).Updates(&Tpl{
		Name:              opts.Name,
		Language:          languages,
		Timeout:           opts.Timeout,
		MaxCPUs:           opts.MaxCPUs,
		MaxMemory:         opts.MaxMemory,
		InternetAccess:    opts.InternetAccess,
		DNS:               dns,
		MaxContainer:      opts.MaxContainer,
		MaxContainerPerIP: opts.MaxContainerPerIP,
	}).Error
}

func (db *tpls) Delete(id uint) error {
	return db.DB.Delete(&Tpl{}, "id = ?", id).Error
}
