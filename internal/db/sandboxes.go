package db

import (
	"errors"

	"github.com/thanhpk/randstr"
	"gorm.io/gorm"
)

type SandboxStore interface {
	// GetByID returns a sandbox with the given id.
	GetByID(id uint) (*Sandbox, error)
	// GetByUID returns a sandbox with the given uid.
	GetByUID(uid string) (*Sandbox, error)
	// ListAll returns all the sandboxes.
	ListAll() ([]*Sandbox, error)
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

func (db *sandboxes) GetByID(id uint) (*Sandbox, error) {
	var sb Sandbox
	return &sb, db.Preload("Template").Model(&Sandbox{}).Where("id = ?", id).First(&sb).Error
}

func (db *sandboxes) GetByUID(uid string) (*Sandbox, error) {
	var sb Sandbox
	return &sb, db.Preload("Template").Model(&Sandbox{}).Where("uid = ?", uid).First(&sb).Error
}

func (db *sandboxes) ListAll() ([]*Sandbox, error) {
	var sbs []*Sandbox
	return sbs, db.Preload("Template").Model(&Sandbox{}).Find(&sbs).Error
}

type CreateSandboxOptions struct {
	Name        string
	TemplateID  uint
	Placeholder string
	Editable    bool
}

func (db *sandboxes) Create(opts CreateSandboxOptions) error {
	return db.DB.Create(&Sandbox{
		UID:         randstr.String(10),
		Name:        opts.Name,
		TemplateID:  opts.TemplateID,
		Placeholder: opts.Placeholder,
		Editable:    opts.Editable,
	}).Error
}

type UpdateSandboxOptions struct {
	ID          uint
	Name        string
	TemplateID  uint
	Placeholder string
	Editable    bool
}

func (db *sandboxes) Update(opts UpdateSandboxOptions) error {
	_, err := db.GetByID(opts.ID)
	if err != nil {
		return errors.New("sandbox not existed")
	}

	return db.DB.Model(&Sandbox{}).Where("id = ?", opts.ID).Updates(&Sandbox{
		Name:        opts.Name,
		TemplateID:  opts.TemplateID,
		Placeholder: opts.Placeholder,
		Editable:    opts.Editable,
	}).Error
}

func (db *sandboxes) Delete(id uint) error {
	return db.DB.Delete(&Sandbox{}, "id = ?", id).Error
}
