// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"testing"

	"github.com/wuhan005/Elaina/internal/config"
)

func TestSandboxSourceLimit(t *testing.T) {
	previous := config.App
	t.Cleanup(func() { config.App = previous })
	config.App.WASMLimits.MaxSourceBytes = 6
	for _, value := range []string{"", "中文", "中文a"} {
		create := CreateSandbox{
			Placeholder: value,
		}
		for _, err := range []error{create.Validate(), UpdateSandbox(create).Validate()} {
			if (err != nil) != (len(value) > 6) {
				t.Fatalf("source %q: %v", value, err)
			}
		}
	}
}
