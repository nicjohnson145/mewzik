package internal

import (
	"bytes"
	"fmt"
	"io/fs"

	"errors"
	"github.com/dhowden/tag"
	"path/filepath"
	"regexp"
)

var (
	multipleUnderscoresRegex = regexp.MustCompile(`_{2,}`)
	ErrMissingMetadataError  = errors.New("missing metadata")
)

func computeDestinationPath(fsys fs.FS, overrides MetadataOverride, path string) (string, error) {
	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	metadata, err := tag.ReadFrom(bytes.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("error parsing metadata: %w", err)
	}

	if metadata.Artist() == "" && overrides.Artist == "" && metadata.AlbumArtist() == "" {
		return "", fmt.Errorf("%w: artist", ErrMissingMetadataError)
	}

	if metadata.Album() == "" && overrides.Album == "" {
		return "", fmt.Errorf("%w: album", ErrMissingMetadataError)
	}

	if metadata.Title() == "" {
		return "", fmt.Errorf("%w: title", ErrMissingMetadataError)
	}

	track, _ := metadata.Track()
	if track == 0 {
		return "", fmt.Errorf("%w: track", ErrMissingMetadataError)
	}

	artist := overrides.Artist
	if artist == "" {
		if metadata.AlbumArtist() != "" {
			artist = metadata.AlbumArtist()
		} else {
			artist = metadata.Artist()
		}
	}

	album := overrides.Album
	if album == "" {
		album = metadata.Album()
	}

	return filepath.Join(
		escape(artist),
		escape(album),
		fmt.Sprintf("%02d_%v", track, escape(metadata.Title())),
	) + filepath.Ext(path), nil
}

func escape(s string) string {
	runes := []rune{}
	for _, r := range s {
		switch r {
		case '"', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '<', '>', '?', ':', '{', '}', '[', ']', '\'', '-', ',', '.':
			continue
		case ' ':
			runes = append(runes, '_')
		default:
			runes = append(runes, r)
		}
	}
	return multipleUnderscoresRegex.ReplaceAllString(string(runes), "_")
}
