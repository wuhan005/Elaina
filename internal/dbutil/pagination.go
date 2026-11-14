// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbutil

const DefaultPageSize = 20

type Pagination struct {
	Page     int
	PageSize int
}

// LimitOffset returns LIMIT and OFFSET parameter for SQL.
// Pages start at 1; non-positive values fall back to the first page and DefaultPageSize.
func (p Pagination) LimitOffset() (limit, offset int) {
	page, pageSize := p.Page, p.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	return pageSize, (page - 1) * pageSize
}
