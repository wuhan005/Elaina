// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"

	"github.com/wuhan005/Elaina/frontend"
	"github.com/wuhan005/Elaina/internal/context"
)

func Frontend(c context.Context) {
	if c.Request().Method != http.MethodGet && c.Request().Method != http.MethodHead {
		return
	}

	name := "index.html"
	f, err := http.FS(frontend.FS).Open(name)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	fi, err := f.Stat()
	if err != nil {
		return // File exists but failed to open.
	}

	http.ServeContent(c.ResponseWriter(), c.Request().Request, name, fi.ModTime(), f)
}
