package cataloger

import (
	"time"
)

type Options struct {
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
