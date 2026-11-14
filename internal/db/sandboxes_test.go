// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"testing"

	"github.com/pkg/errors"
)

func TestSandboxUpdateAndDelete(t *testing.T) {
	newTestDatabase(t)
	ctx := context.Background()
	var templates []*Tpl
	for _, name := range []string{"First", "Second"} {
		template, err := Tpls.Create(ctx, CreateTplOptions{
			Name:     name,
			Language: []string{"python"},
		})
		if err != nil {
			t.Fatal(err)
		}
		templates = append(templates, template)
	}

	sandbox, err := Sandboxes.Create(ctx, CreateSandboxOptions{
		Name:        "Sandbox",
		TemplateID:  templates[0].ID,
		Placeholder: "print(1)",
		Editable:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := Sandboxes.Update(ctx, sandbox.ID, UpdateSandboxOptions{
		Name:       "Updated",
		TemplateID: templates[1].ID,
	}); err != nil {
		t.Fatal(err)
	}
	stored, err := Sandboxes.GetByID(ctx, sandbox.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != "Updated" || stored.TemplateID != templates[1].ID || stored.Template.ID != templates[1].ID ||
		stored.Placeholder != "" || stored.Editable || stored.UID != sandbox.UID || !stored.UpdatedAt.After(sandbox.UpdatedAt) {
		t.Fatalf("sandbox was not updated: %+v", stored)
	}
	if stored.Template.Name != "Second" {
		t.Fatalf("sandbox update changed the associated template: %+v", stored.Template)
	}

	if err := Sandboxes.Update(ctx, sandbox.ID+1000, UpdateSandboxOptions{
		Name:       "Missing",
		TemplateID: templates[0].ID,
	}); !errors.Is(err, ErrSandboxNotFound) {
		t.Fatalf("expected ErrSandboxNotFound when updating, got %v", err)
	}
	if err := Sandboxes.Delete(ctx, sandbox.ID); err != nil {
		t.Fatal(err)
	}
	if err := Sandboxes.Delete(ctx, sandbox.ID); !errors.Is(err, ErrSandboxNotFound) {
		t.Fatalf("expected ErrSandboxNotFound when deleting twice, got %v", err)
	}
	if err := Tpls.Delete(ctx, templates[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := Tpls.Delete(ctx, templates[0].ID); !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("expected ErrTemplateNotFound when deleting twice, got %v", err)
	}
	if err := Tpls.Update(ctx, templates[0].ID, UpdateTplOptions{
		Name: "Deleted",
	}); !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("expected ErrTemplateNotFound when updating a deleted template, got %v", err)
	}
}
