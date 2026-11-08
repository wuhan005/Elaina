package db

import (
	"context"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/dbutil"
)

var _ TplStore = (*tpls)(nil)

var Tpls TplStore

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

func NewTplStore(db *gorm.DB) TplStore {
	return &tpls{
		db,
	}
}

type tpls struct {
	*gorm.DB
}

type Tpl struct {
	dbutil.Model

	Name     string         `json:"name"`
	Language pq.StringArray `gorm:"type:text[]" json:"language"`

	Timeout        int   `json:"timeout"`
	MaxMemory      int64 `json:"maxMemory"`
	InternetAccess bool  `json:"internetAccess"`

	// Retain legacy CPU quotas until the user saves the template without them.
	MaxCPUs int64 `json:"maxCpus"`
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
	Name           string
	Language       []string
	Timeout        int
	MaxMemory      int64
	InternetAccess bool
}

func (db *tpls) Create(ctx context.Context, options CreateTplOptions) (*Tpl, error) {
	tpl := &Tpl{
		Name:           options.Name,
		Language:       pq.StringArray(options.Language),
		Timeout:        options.Timeout,
		MaxMemory:      options.MaxMemory,
		InternetAccess: options.InternetAccess,
	}

	if err := db.WithContext(ctx).Create(tpl).Error; err != nil {
		return nil, errors.Wrap(err, "create")
	}
	return tpl, nil
}

type UpdateTplOptions struct {
	ID             uint
	Name           string
	Language       []string
	Timeout        int
	MaxMemory      int64
	InternetAccess bool
}

func (db *tpls) Update(ctx context.Context, id uint, options UpdateTplOptions) error {
	template, err := db.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get by ID")
	}

	template.Name = options.Name
	template.Language = options.Language
	template.Timeout = options.Timeout
	template.MaxCPUs = 0
	template.MaxMemory = options.MaxMemory
	template.InternetAccess = options.InternetAccess

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
