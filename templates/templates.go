package templates

import "embed"

//go:embed index.html history.html settings.html
var FS embed.FS
