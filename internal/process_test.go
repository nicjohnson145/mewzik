package internal

import (
	"path/filepath"
	"testing"

	"github.com/deckarep/golang-set"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/require"
)

func mustTouch(t *testing.T, fsys *memfs.FS, path string) {
	require.NoError(t, fsys.MkdirAll(filepath.Dir(path), 0775))
	require.NoError(t, fsys.WriteFile(path, []byte("foo"), 0664))
}

func TestGetFileList(t *testing.T) {
	t.Run("smokes", func(t *testing.T) {
		rootFS := memfs.New()
		mustTouch(t, rootFS, "artist1/album 1/song1.mp3")
		mustTouch(t, rootFS, "artist1/album 1/song2.mp3")
		mustTouch(t, rootFS, "artist2/album 3/song1.mp3")
		mustTouch(t, rootFS, "artist2/album 3/song2.mp3")

		proc := NewProcessor(ProcessorConfig{
			Extensions: mapset.NewSet(".mp3"),
		})

		got, err := proc.getFileList(rootFS)
		require.NoError(t, err)
		require.Equal(
			t,
			[]string{
				"artist1/album 1/song1.mp3",
				"artist1/album 1/song2.mp3",
				"artist2/album 3/song1.mp3",
				"artist2/album 3/song2.mp3",
			},
			got,
		)
	})
}
