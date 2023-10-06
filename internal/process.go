package internal

import (
	"github.com/deckarep/golang-set"
	"github.com/jarxorg/wfs"
	"github.com/rs/zerolog"
	"io/fs"
	"path/filepath"
)

type MetadataOverride struct {
	Artist string
	Album  string
}

type ProcessorConfig struct {
	Logger     zerolog.Logger
	InputFS    fs.FS
	OutputFS   wfs.WriteFileFS
	Extensions mapset.Set
	Overrides  MetadataOverride
}

func NewProcessor(conf ProcessorConfig) *Processor {
	return &Processor{
		log:        conf.Logger,
		inputFS:    conf.InputFS,
		outputFS:   conf.OutputFS,
		extensions: conf.Extensions,
		overrides:  conf.Overrides,
	}
}

type Processor struct {
	log        zerolog.Logger
	inputFS    fs.FS
	outputFS   wfs.WriteFileFS
	extensions mapset.Set
	overrides  MetadataOverride
}

func (p *Processor) Process() error {
	inputFiles, err := p.getFileList(p.inputFS)
	if err != nil {
		return err
	}

	mappings, err := p.getMapping(p.inputFS, inputFiles)
	if err != nil {
		return err
	}

	if err := p.executeCopy(p.inputFS, p.outputFS, mappings); err != nil {
		return err
	}

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

func (p *Processor) getMapping(fsys fs.FS, files []string) (map[string]string, error) {
	mapping := map[string]string{}

	for _, fl := range files {
		destPath, err := computeDestinationPath(fsys, p.overrides, fl)
		if err != nil {
			return nil, err
		}

		mapping[fl] = destPath
	}

	return mapping, nil
}

func (p *Processor) executeCopy(input fs.FS, output wfs.WriteFileFS, mapping map[string]string) error {
	for inputPath, outputPath := range mapping {
		containing := filepath.Dir(outputPath)

		// TODO: make perms configurable
		if err := output.MkdirAll(containing, 0755); err != nil {
			p.log.Err(err).Str("path", outputPath).Msg("error making containing directory")
			return err
		}

		content, err := fs.ReadFile(input, inputPath)
		if err != nil {
			p.log.Err(err).Str("path", inputPath).Msg("error reading source file")
			return err
		}

		// TODO: make perms configurable
		if _, err := wfs.WriteFile(output, outputPath, content, 0644); err != nil {
			p.log.Err(err).Str("path", outputPath).Msg("error writing output")
			return err
		}
	}

	return nil
}
