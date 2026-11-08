// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbutil

import "time"

func Now() time.Time {
	return time.Now().Truncate(time.Microsecond)
}
