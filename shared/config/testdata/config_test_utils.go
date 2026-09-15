package testdata

import "embed"

//go:embed base.yaml test.yaml
var Files embed.FS
