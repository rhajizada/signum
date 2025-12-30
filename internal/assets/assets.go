package assets

import (
	"embed"
	"io/fs"
)

//go:embed files/favicon/*.png files/logo/*.png
var embeddedFiles embed.FS

// Files provides the embedded assets root for the HTTP file server.
func Files() fs.FS {
	return mustSub(embeddedFiles, "files")
}

func mustSub(fsys embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
