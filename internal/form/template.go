// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

type CreateTemplate struct {
	Name              string            `json:"name" valid:"required"`
	Language          []string          `json:"language" valid:"required"`
	Timeout           int               `json:"timeout" valid:"required;min:0;max:60"`
	MaxCPUs           int64             `json:"maxCpus" valid:"required;min:0;max:10"`
	MaxMemory         int64             `json:"maxMemory" valid:"required;min:6;max:2048"`
	InternetAccess    bool              `json:"internetAccess"`
	DNS               map[string]string `json:"dns"`
	MaxContainer      int64             `json:"maxContainer" valid:"required;min:0;max:1000"`
	MaxContainerPerIP int64             `json:"maxContainerPerIP" valid:"required;min:0;max:100"`
}

type UpdateTemplate struct {
	Name              string            `json:"name" valid:"required"`
	Language          []string          `json:"language" valid:"required"`
	Timeout           int               `json:"timeout" valid:"required;min:0;max:60"`
	MaxCPUs           int64             `json:"maxCpus" valid:"required;min:0;max:10"`
	MaxMemory         int64             `json:"maxMemory" valid:"required;min:6;max:2048"`
	InternetAccess    bool              `json:"internetAccess"`
	DNS               map[string]string `json:"dns"`
	MaxContainer      int64             `json:"maxContainer" valid:"required;min:0;max:1000"`
	MaxContainerPerIP int64             `json:"maxContainerPerIP" valid:"required;min:0;max:100"`
}
