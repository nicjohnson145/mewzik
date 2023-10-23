package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/deckarep/golang-set/v2"
	"github.com/jarxorg/wfs/osfs"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/require"
	"sort"
	"github.com/dhowden/tag"
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
		mustTouch(t, rootFS, "artist1/album 1/cover.jpg")
		mustTouch(t, rootFS, "artist2/album 3/song1.mp3")
		mustTouch(t, rootFS, "artist2/album 3/song2.mp3")
		mustTouch(t, rootFS, "artist2/album 3/README.rst")

		proc, err := NewProcessor(ProcessorConfig{
			Extensions: mapset.NewSet(".mp3"),
		})
		require.NoError(t, err)

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

func TestProcess(t *testing.T) {
	sortedFileList := func(t *testing.T, fsys fs.FS) []string {
		t.Helper()

		files := []string{}
		require.NoError(t, fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, e1 error) error {
			if e1 != nil {
				return e1
			}

			if d.IsDir() {
				return nil
			}

			files = append(files, path)
			return nil
		}))

		sort.Strings(files)

		return files
	}

	t.Run("smokes", func(t *testing.T) {
		dir := t.TempDir()

		input := os.DirFS("./testdata/sample-files/Anders - 669")
		output := osfs.New(dir)

		p, err := NewProcessor(ProcessorConfig{
			InputFS: input,
			OutputFS: output,
			Extensions: mapset.NewSet(".mp3"),
		})
		require.NoError(t, err)

		require.NoError(t, p.Process())
		files := sortedFileList(t, output)

		require.Equal(
			t,
			[]string{
				"Anders/669/01_With_or_Without.mp3",
				"Anders/669/02_Diamonds.mp3",
			},
			files,
		)
	})

	t.Run("with overrides", func(t *testing.T) {
		dir := t.TempDir()

		input := os.DirFS("./testdata/sample-files/Anders - 669")
		output := osfs.New(dir)

		p, err := NewProcessor(ProcessorConfig{
			InputFS: input,
			OutputFS: output,
			Extensions: mapset.NewSet(".mp3"),
			OutputRoot: dir,
			Overrides: MetadataOverride{
				Album: "New Album",
				Artist: "New Artist",
			},
		})
		require.NoError(t, err)
		require.NoError(t, p.Process())
		require.Equal(
			t,
			[]string{
				"New_Artist/New_Album/01_With_or_Without.mp3",
				"New_Artist/New_Album/02_Diamonds.mp3",
			},
			sortedFileList(t, output),
		)

		// Also the ID3 tags on the files should be updated as well
		f, err := os.Open(filepath.Join(dir, "New_Artist/New_Album/01_With_or_Without.mp3"))
		require.NoError(t, err)
		defer f.Close()

		md, err := tag.ReadFrom(f)
		require.NoError(t, err)

		require.Equal(t, "New Artist", md.Artist())
		require.Equal(t, "New Album", md.Album())
	})
}
