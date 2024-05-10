package public

import (
	"embed"
)

//go:embed "html" "static"
var Files embed.FS
