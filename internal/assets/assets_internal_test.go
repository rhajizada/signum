package assets

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustSub(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantPanic bool
	}{
		{name: "returns sub fs", path: "files"},
		{name: "panics on missing path", path: "/missing", wantPanic: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				assert.Panics(t, func() {
					_ = mustSub(embeddedFiles, tc.path)
				})
				return
			}

			sub := mustSub(embeddedFiles, tc.path)
			_, err := fs.Stat(sub, "logo/signum.png")
			require.NoError(t, err)
		})
	}
}
