package public

import (
	"embed"
)

//go:embed css/* js/*
var FS embed.FS
