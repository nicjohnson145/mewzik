package internal

import (
	"github.com/deckarep/golang-set"
	"github.com/rs/zerolog"
	"io/fs"
	"path/filepath"
)

type ProcessorConfig struct {
	Logger     zerolog.Logger
	InputFS    fs.FS
	OutputFS   fs.FS
	Extensions mapset.Set
}

func NewProcessor(conf ProcessorConfig) *Processor {
	return &Processor{
		log:        conf.Logger,
		inputFS:    conf.InputFS,
		outputFS:   conf.OutputFS,
		extensions: conf.Extensions,
	}
}

type Processor struct {
	log        zerolog.Logger
	inputFS    fs.FS
	outputFS   fs.FS
	extensions mapset.Set
}

func (p *Processor) Process() error {
	return nil
}

func (p *Processor) getFileList(fsys fs.FS) ([]string, error) {
	foundFiles := []string{}
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !p.extensions.Contains(filepath.Ext(path)) {
			p.log.Debug().Str("path", path).Msg("skipping due to extension filter")
			return nil
		}

		foundFiles = append(foundFiles, path)

		return nil
	})
	return foundFiles, nil
}
