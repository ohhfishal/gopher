package testdata

import "embed"

//go:embed *
var FS embed.FS

//go:embed buildOutputs*
var BuildOutputs embed.FS
