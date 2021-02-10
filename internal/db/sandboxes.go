package db

import (
	"errors"

	"github.com/thanhpk/randstr"
	"gorm.io/gorm"
)

type SandboxStore interface {
	// GetByID returns a sandbox with the given id.
	GetByID(id uint) (*sandbox, error)
	// ListAll returns all the sandboxes.
	ListAll() ([]*sandbox, error)
	// Create creates a new sandbox with the given options.
	Create(opts CreateSandboxOptions) error
	// Update edits a new sandbox with the given options.
	Update(opts UpdateSandboxOptions) error
	// Delete deletes a sandbox with the given id.
	Delete(id uint) error
}

var Sandboxes SandboxStore

var _ SandboxStore = (*sandboxes)(nil)

type sandboxes struct {
	*gorm.DB
}

func (db *sandboxes) GetByID(id uint) (*sandbox, error) {
	var sb sandbox
	return &sb, db.Preload("Template").Model(&sandbox{}).Where("id = ?", id).First(&sb).Error
}

func (db *sandboxes) ListAll() ([]*sandbox, error) {
	var sbs []*sandbox
	return sbs, db.Preload("Template").Model(&sandbox{}).Find(&sbs).Error
}

type CreateSandboxOptions struct {
	TemplateID  uint
	Placeholder string
	Editable    bool
}

func (db *sandboxes) Create(opts CreateSandboxOptions) error {
	return db.DB.Create(&sandbox{
		UID:         randstr.String(10),
		TemplateID:  opts.TemplateID,
		Placeholder: opts.Placeholder,
		Editable:    opts.Editable,
	}).Error
}

type UpdateSandboxOptions struct {
	ID          uint
	TemplateID  uint
	Placeholder string
	Editable    bool
}

func (db *sandboxes) Update(opts UpdateSandboxOptions) error {
	_, err := db.GetByID(opts.ID)
	if err != nil {
		return errors.New("sandbox not existed")
	}

	return db.DB.Model(&sandbox{}).Where(&sandbox{
		Model: gorm.Model{
			ID: opts.ID,
		},
	}).Updates(&sandbox{
		TemplateID:  opts.TemplateID,
		Placeholder: opts.Placeholder,
		Editable:    opts.Editable,
	}).Error
}

func (db *sandboxes) Delete(id uint) error {
	return db.DB.Delete(&sandboxes{}, "id = ?", id).Error
}
