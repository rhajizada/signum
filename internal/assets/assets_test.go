package assets_test

import (
	"io/fs"
	"testing"

	"github.com/rhajizada/signum/internal/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesContainAssets(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "favicon", path: "favicon/favicon.png"},
		{name: "logo", path: "logo/signum.png"},
	}

	assetFS := assets.Files()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, err := fs.Stat(assetFS, tc.path)
			require.NoError(t, err)
			assert.Positive(t, info.Size())
		})
	}
}
