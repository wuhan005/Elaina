package db

import (
	"context"

	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/dbutil"
)

var _ SandboxStore = (*sandboxes)(nil)

var Sandboxes SandboxStore

type SandboxStore interface {
	// List returns the sandboxes with the given options.
	List(ctx context.Context, options ListSandboxOptions) ([]*Sandbox, int64, error)
	// GetByID returns a sandbox with the given id.
	GetByID(ctx context.Context, id uint) (*Sandbox, error)
	// GetByUID returns a sandbox with the given uid.
	GetByUID(ctx context.Context, uid string) (*Sandbox, error)
	// Create creates a new sandbox with the given options.
	Create(ctx context.Context, options CreateSandboxOptions) (*Sandbox, error)
	// Update edits a sandbox with the given options.
	Update(ctx context.Context, id uint, options UpdateSandboxOptions) error
	// Delete deletes a sandbox with the given id.
	Delete(ctx context.Context, id uint) error
}

func NewSandboxStore(db *gorm.DB) SandboxStore {
	return &sandboxes{
		db,
	}
}

type sandboxes struct {
	*gorm.DB
}

type Sandbox struct {
	dbutil.Model

	UID         string `gorm:"NOT NULL"`
	Name        string
	TemplateID  uint `gorm:"NOT NULL"`
	Template    *Tpl `gorm:"ForeignKey:TemplateID"`
	Placeholder string
	Editable    bool
}

var ErrSandboxNotFound = errors.New("sandbox does not exist")

func (db *sandboxes) getBy(ctx context.Context, query string, args ...any) (*Sandbox, error) {
	var sandbox Sandbox
	if err := db.WithContext(ctx).Preload("Template").Where(query, args...).First(&sandbox).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSandboxNotFound
		}
		return nil, errors.Wrap(err, "first")
	}
	return &sandbox, nil
}

func (db *sandboxes) GetByID(ctx context.Context, id uint) (*Sandbox, error) {
	return db.getBy(ctx, "id = ?", id)
}

func (db *sandboxes) GetByUID(ctx context.Context, uid string) (*Sandbox, error) {
	return db.getBy(ctx, "uid = ?", uid)
}

type ListSandboxOptions struct {
	dbutil.Pagination
}

func (db *sandboxes) List(ctx context.Context, options ListSandboxOptions) ([]*Sandbox, int64, error) {
	var sandboxes []*Sandbox

	query := db.WithContext(ctx).Model(&Sandbox{}).Preload("Template")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, "count")
	}

	limit, offset := options.LimitOffset()
	if err := query.Limit(limit).Offset(offset).Find(&sandboxes).Error; err != nil {
		return nil, 0, errors.Wrap(err, "find")
	}
	return sandboxes, total, nil
}

type CreateSandboxOptions struct {
	Name        string
	TemplateID  uint
	Placeholder string
	Editable    bool
}

func (db *sandboxes) Create(ctx context.Context, options CreateSandboxOptions) (*Sandbox, error) {
	sandbox := &Sandbox{
		UID:         randstr.String(10),
		Name:        options.Name,
		TemplateID:  options.TemplateID,
		Placeholder: options.Placeholder,
		Editable:    options.Editable,
	}

	if err := db.WithContext(ctx).Create(sandbox).Error; err != nil {
		return nil, errors.Wrap(err, "create")
	}
	return sandbox, nil
}

type UpdateSandboxOptions struct {
	Name        string
	TemplateID  uint
	Placeholder string
	Editable    bool
}

func (db *sandboxes) Update(ctx context.Context, id uint, options UpdateSandboxOptions) error {
	// Select explicitly so zero values are written too.
	result := db.WithContext(ctx).Model(&Sandbox{}).Where("id = ?", id).
		Select("Name", "TemplateID", "Placeholder", "Editable").
		Updates(&Sandbox{
			Name:        options.Name,
			TemplateID:  options.TemplateID,
			Placeholder: options.Placeholder,
			Editable:    options.Editable,
		})
	if result.Error != nil {
		return errors.Wrap(result.Error, "update")
	}
	if result.RowsAffected == 0 {
		return ErrSandboxNotFound
	}
	return nil
}

func (db *sandboxes) Delete(ctx context.Context, id uint) error {
	result := db.WithContext(ctx).Delete(&Sandbox{}, "id = ?", id)
	if result.Error != nil {
		return errors.Wrap(result.Error, "delete")
	}
	if result.RowsAffected == 0 {
		return ErrSandboxNotFound
	}
	return nil
}
