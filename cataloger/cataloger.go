package cataloger

import (
	"io/fs"
	"time"
)

type Options struct {
	RootDir         string
	FS              fs.FS
	AnchorFile      string
	IgnoreFile      string
	OutputDir       string
	BaseDirectories []string
	TimeLocation    *time.Location
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
