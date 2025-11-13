package web

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distFS embed.FS

// GetDistFS returns the embedded web dist filesystem
func GetDistFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
