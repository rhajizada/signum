package assets

import (
	"io/fs"
	"testing"
)

func TestMustSubReturnsFS(t *testing.T) {
	sub := mustSub(embeddedFiles, "files")
	if _, err := fs.Stat(sub, "logo/signum.png"); err != nil {
		t.Fatalf("expected embedded asset: %v", err)
	}
}

func TestMustSubPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic for missing subdirectory")
		}
	}()
	_ = mustSub(embeddedFiles, "/missing")
}
