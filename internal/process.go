package internal

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/bogem/id3v2/v2"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/jarxorg/wfs"
	"github.com/rs/zerolog"
)

type MetadataOverride struct {
	Artist      string
	Album       string
	AlbumArtist string
}

func (m *MetadataOverride) active() bool {
	return m.Artist != "" || m.Album != "" || m.AlbumArtist != ""
}

type ProcessorConfig struct {
	Logger     zerolog.Logger
	InputFS    fs.FS
	OutputFS   wfs.WriteFileFS
	OutputRoot string
	Extensions mapset.Set[string]
	Overrides  MetadataOverride
}

func NewProcessor(conf ProcessorConfig) (*Processor, error) {
	if conf.Overrides.active() && conf.OutputRoot == "" {
		return nil, fmt.Errorf("must supply output root if overrides in use")
	}
	return &Processor{
		log:        conf.Logger,
		inputFS:    conf.InputFS,
		outputFS:   conf.OutputFS,
		outputRoot: conf.OutputRoot,
		extensions: conf.Extensions,
		overrides:  conf.Overrides,
	}, nil
}

type Processor struct {
	log        zerolog.Logger
	inputFS    fs.FS
	outputFS   wfs.WriteFileFS
	outputRoot string
	extensions mapset.Set[string]
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

	if p.overrides.active() {
		if err := p.overwriteTags(mappings); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) getFileList(fsys fs.FS) ([]string, error) {
	foundFiles := []string{}
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
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
	if err != nil {
		return nil, err
	}
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

func (p *Processor) overwriteTags(mapping map[string]string) error {
	for _, outputFragment := range mapping {
		outputPath := filepath.Join(p.outputRoot, outputFragment)
		if err := p.overwriteSingleTag(outputPath); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) overwriteSingleTag(path string) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		p.log.Err(err).Msg("error opening destination path for tag writing")
		return err
	}
	defer tag.Close()

	if p.overrides.Artist != "" {
		tag.SetArtist(p.overrides.Artist)
	}

	if p.overrides.Album != "" {
		tag.SetAlbum(p.overrides.Album)
	}

	if p.overrides.AlbumArtist != "" {
		tag.AddTextFrame(tag.CommonID("Band/Orchestra/Accompaniment"), tag.DefaultEncoding(), p.overrides.AlbumArtist)
	}

	if err := tag.Save(); err != nil {
		p.log.Err(err).Msg("error saving tag update")
		return err
	}

	return nil
}
