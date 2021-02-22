package public

import (
	"embed"
)

//go:embed css/* js/* assets/*
var FS embed.FS
