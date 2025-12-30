package assets_test

import (
	"io/fs"
	"testing"

	"github.com/rhajizada/signum/internal/assets"
)

func TestFilesContainAssets(t *testing.T) {
	assetFS := assets.Files()
	paths := []string{
		"favicon/favicon.png",
		"logo/signum.png",
	}
	for _, path := range paths {
		info, err := fs.Stat(assetFS, path)
		if err != nil {
			t.Fatalf("expected asset %s: %v", path, err)
		}
		if info.Size() == 0 {
			t.Fatalf("expected asset %s to have content", path)
		}
	}
}
