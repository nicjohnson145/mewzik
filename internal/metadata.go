package internal

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/dhowden/tag"
	"path/filepath"
	"regexp"
	"errors"
)

var multipleUnderscoresRegex = regexp.MustCompile(`_{2,}`)

var ErrMissingMetadataError = errors.New("missing metadata")

func computeDestinationPath(fsys fs.FS, path string) (string, error) {
	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	metadata, err := tag.ReadFrom(bytes.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("error parsing metadata: %w", err)
	}

	if metadata.Artist() == "" {
		return "", fmt.Errorf("%w: artist", ErrMissingMetadataError)
	}

	if metadata.Album() == "" {
		return "", fmt.Errorf("%w: album", ErrMissingMetadataError)
	}

	if metadata.Title() == "" {
		return "", fmt.Errorf("%w: title", ErrMissingMetadataError)
	}

	track, _ := metadata.Track()
	if track == 0 {
		return "", fmt.Errorf("%w: track", ErrMissingMetadataError)
	}

	return filepath.Join(
		escape(metadata.Artist()),
		escape(metadata.Album()),
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
