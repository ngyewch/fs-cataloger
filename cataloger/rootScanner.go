package cataloger

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"
)

type rootScanner struct {
	options          Options
	ignoredWriter    io.WriteCloser
	unfiledWriter    io.WriteCloser
	unfiledCsvWriter *csv.Writer
}

func (s *rootScanner) Close() error {
	if s.ignoredWriter != nil {
		err := s.ignoredWriter.Close()
		if err != nil {
			return err
		}
	}
	if s.unfiledCsvWriter != nil {
		s.unfiledCsvWriter.Flush()
	}
	if s.unfiledWriter != nil {
		err := s.unfiledWriter.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *rootScanner) recordIgnored(path string) error {
	if s.ignoredWriter == nil {
		f, err := os.Create(filepath.Join(s.options.OutputDir, "ignored.txt"))
		if err != nil {
			return err
		}
		s.ignoredWriter = f
	}
	_, err := io.WriteString(s.ignoredWriter, fmt.Sprintf("%s\n", path))
	return err
}

func (s *rootScanner) recordUnfiled(path string, d fs.DirEntry) error {
	if s.unfiledCsvWriter == nil {
		if s.unfiledWriter == nil {
			f, err := os.Create(filepath.Join(s.options.OutputDir, "unfiled.csv"))
			if err != nil {
				return err
			}
			s.unfiledWriter = f
		}
		s.unfiledCsvWriter = csv.NewWriter(s.unfiledWriter)
		err := s.unfiledCsvWriter.Write([]string{"path", "size", "modified"})
		if err != nil {
			return err
		}
	}
	fileInfo, err := d.Info()
	if err != nil {
		return err
	}
	return s.unfiledCsvWriter.Write([]string{
		path,
		fmt.Sprintf("%d", fileInfo.Size()),
		fileInfo.ModTime().Format(time.RFC3339),
	})
}

func (s *rootScanner) generate() error {
	err := os.RemoveAll(s.options.OutputDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(s.options.OutputDir, 0755)
	if err != nil {
		return err
	}

	for _, source := range s.options.Sources {
		err = s.processSource(source)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *rootScanner) processSource(source Source) error {
	return fs.WalkDir(source.FS, ".", func(path1 string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			ignoreFile := path.Join(path1, s.options.IgnoreFile)
			ignoreFileStat, err := fs.Stat(source.FS, ignoreFile)
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else if !ignoreFileStat.IsDir() {
				err = s.recordIgnored(path1 + "/")
				if err != nil {
					return err
				}
				return fs.SkipDir
			}

			anchorFile := path.Join(path1, s.options.AnchorFile)
			anchorFileStat, err := fs.Stat(source.FS, anchorFile)
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else if !anchorFileStat.IsDir() {
				subFs, err := fs.Sub(source.FS, path1)
				if err != nil {
					return err
				}
				ps, err := newProjectScanner(subFs, path.Join(source.RootDir, path1), filepath.Join(s.options.OutputDir, d.Name()), s.options.IgnoreFile)
				if err != nil {
					return err
				}
				err = ps.generate()
				if err != nil {
					return err
				}
				err = ps.Close()
				if err != nil {
					return err
				}
				return fs.SkipDir
			}
		} else {
			return s.recordUnfiled(path1, d)
		}
		return nil
	})
}
