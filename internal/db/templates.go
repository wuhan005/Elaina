package db

import (
	"context"
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/dbutil"
)

type TplStore interface {
	// All returns all the templates.
	All(ctx context.Context) ([]*Tpl, error)
	// List returns the templates with the given options.
	List(ctx context.Context, options ListTplOptions) ([]*Tpl, int64, error)
	// GetByID returns a template with the given id.
	GetByID(ctx context.Context, id uint) (*Tpl, error)
	// Create creates a new template with the given options.
	Create(ctx context.Context, options CreateTplOptions) (*Tpl, error)
	// Update edits a new template with the given options.
	Update(ctx context.Context, id uint, options UpdateTplOptions) error
	// Delete deletes a template with the given id.
	Delete(ctx context.Context, id uint) error
}

var Tpls TplStore

var _ TplStore = (*tpls)(nil)

type tpls struct {
	*gorm.DB
}

type Tpl struct {
	dbutil.Model

	Name     string         `json:"name"`
	Language pq.StringArray `gorm:"type:text[]" json:"language"`

	// Limit
	Timeout           int            `json:"timeout"`
	MaxCPUs           int64          `json:"maxCpus"`
	MaxMemory         int64          `json:"maxMemory"`
	InternetAccess    bool           `json:"internetAccess"`
	DNS               datatypes.JSON `gorm:"type:jsonb" json:"dns"`
	MaxContainer      int64          `json:"maxContainer"`
	MaxContainerPerIP int64          `json:"maxContainerPerIp"`
}

var ErrTemplateNotFound = errors.New("template dose not exist")

func (db *tpls) GetByID(ctx context.Context, id uint) (*Tpl, error) {
	var template Tpl

	if err := db.WithContext(ctx).Model(&Tpl{}).Where("id = ?", id).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}

		return nil, errors.Wrap(err, "first")
	}
	return &template, nil
}

func (db *tpls) All(ctx context.Context) ([]*Tpl, error) {
	var templates []*Tpl
	if err := db.WithContext(ctx).Model(&Tpl{}).Find(&templates).Error; err != nil {
		return nil, errors.Wrap(err, "find")
	}
	return templates, nil
}

type ListTplOptions struct {
	dbutil.Pagination
}

func (db *tpls) List(ctx context.Context, options ListTplOptions) ([]*Tpl, int64, error) {
	var templates []*Tpl

	query := db.WithContext(ctx).Model(&Tpl{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, "count")
	}

	limit, offset := options.LimitOffset()
	if err := query.Limit(limit).Offset(offset).Find(&templates).Error; err != nil {
		return nil, 0, errors.Wrap(err, "find")
	}
	return templates, total, nil
}

type CreateTplOptions struct {
	Name              string
	Language          []string
	Timeout           int
	MaxCPUs           int64
	MaxMemory         int64
	InternetAccess    bool
	DNS               map[string]string
	MaxContainer      int64
	MaxContainerPerIP int64
}

func (db *tpls) Create(ctx context.Context, options CreateTplOptions) (*Tpl, error) {
	dnsValue, err := json.Marshal(options.DNS)
	if err != nil {
		return nil, errors.Wrap(err, "marshal dns")
	}
	dns := datatypes.JSON{}
	if err := dns.Scan(dnsValue); err != nil {
		return nil, errors.Wrap(err, "marshal DNS JSONs")
	}

	tpl := &Tpl{
		Name:              options.Name,
		Language:          pq.StringArray(options.Language),
		Timeout:           options.Timeout,
		MaxCPUs:           options.MaxCPUs,
		MaxMemory:         options.MaxMemory,
		InternetAccess:    options.InternetAccess,
		DNS:               dns,
		MaxContainer:      options.MaxContainer,
		MaxContainerPerIP: options.MaxContainerPerIP,
	}

	if err := db.WithContext(ctx).Create(tpl).Error; err != nil {
		return nil, errors.Wrap(err, "create")
	}
	return tpl, nil
}

type UpdateTplOptions struct {
	ID                uint
	Name              string
	Language          []string
	Timeout           int
	MaxCPUs           int64
	MaxMemory         int64
	InternetAccess    bool
	DNS               map[string]string
	MaxContainer      int64
	MaxContainerPerIP int64
}

func (db *tpls) Update(ctx context.Context, id uint, options UpdateTplOptions) error {
	template, err := db.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get by ID")
	}

	dnsValue, err := jsoniter.Marshal(options.DNS)
	if err != nil {
		return errors.Wrap(err, "marshal dns")
	}
	dns := datatypes.JSON{}
	if err := dns.Scan(dnsValue); err != nil {
		return errors.Wrap(err, "marshal DNS JSONs")
	}

	template.Name = options.Name
	template.Language = options.Language
	template.Timeout = options.Timeout
	template.MaxCPUs = options.MaxCPUs
	template.MaxMemory = options.MaxMemory
	template.InternetAccess = options.InternetAccess
	template.DNS = dns
	template.MaxContainer = options.MaxContainer
	template.MaxContainerPerIP = options.MaxContainerPerIP

	if err := db.WithContext(ctx).Save(template).Error; err != nil {
		return errors.Wrap(err, "save")
	}
	return nil
}

func (db *tpls) Delete(ctx context.Context, id uint) error {
	_, err := db.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get by ID")
	}

	if err := db.WithContext(ctx).Delete(&Tpl{}, "id = ?", id).Error; err != nil {
		return errors.Wrap(err, "delete")
	}
	return nil
}
