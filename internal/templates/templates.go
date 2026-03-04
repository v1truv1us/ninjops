package templates

import (
	"embed"
	"io/fs"
)

//go:embed *.tmpl
var templateFS embed.FS

func GetFS() fs.FS {
	return templateFS
}

func ReadTemplate(name string) ([]byte, error) {
	return templateFS.ReadFile(name)
}
