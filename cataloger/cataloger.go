package cataloger

import (
	"io/fs"
)

type Options struct {
	AnchorFile string
	IgnoreFile string
	OutputDir  string
	Sources    []Source
}

type Source struct {
	RootDir string
	FS      fs.FS
}

func Generate(options Options) error {
	s := &rootScanner{
		options: options,
	}
	defer func(s *rootScanner) {
		_ = s.Close()
	}(s)
	return s.generate()
}
