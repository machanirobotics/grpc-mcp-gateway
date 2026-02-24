package templates

import "embed"

//go:embed *.tpl cpp/*.tpl
var FS embed.FS
