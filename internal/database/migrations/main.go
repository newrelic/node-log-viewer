package migrations

import (
	"embed"
	"io/fs"
)

//go:embed sql
var FS embed.FS

func ListFiles() ([]fs.DirEntry, error) {
	return FS.ReadDir("sql")
}
