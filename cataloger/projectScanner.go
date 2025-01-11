package cataloger

import (
	"encoding/csv"
	"fmt"
	"github.com/goccy/go-yaml"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type projectScanner struct {
	fs             fs.FS
	outputDir      string
	ignoreFile     string
	metadata       *metadata
	ignoredWriter  io.WriteCloser
	filedWriter    io.WriteCloser
	filedCsvWriter *csv.Writer
}

type metadata struct {
	Path      string `json:"path"`
	FileCount int64  `json:"fileCount"`
	TotalSize int64  `json:"totalSize"`
}

func newProjectScanner(baseFS fs.FS, baseDir string, outputDir string, ignoreFile string) (*projectScanner, error) {
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return nil, err
	}
	return &projectScanner{
		fs:         baseFS,
		outputDir:  outputDir,
		ignoreFile: ignoreFile,
		metadata: &metadata{
			Path: baseDir,
		},
	}, nil
}

func (s *projectScanner) Close() error {
	if s.ignoredWriter != nil {
		err := s.ignoredWriter.Close()
		if err != nil {
			return err
		}
	}
	if s.filedCsvWriter != nil {
		s.filedCsvWriter.Flush()
	}
	if s.filedWriter != nil {
		err := s.filedWriter.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *projectScanner) recordIgnored(path string) error {
	if s.ignoredWriter == nil {
		f, err := os.Create(filepath.Join(s.outputDir, "ignored.txt"))
		if err != nil {
			return err
		}
		s.ignoredWriter = f
	}
	_, err := io.WriteString(s.ignoredWriter, fmt.Sprintf("%s\n", path))
	return err
}

func (s *projectScanner) recordFiled(path string, d fs.DirEntry) error {
	if s.filedCsvWriter == nil {
		if s.filedWriter == nil {
			f, err := os.Create(filepath.Join(s.outputDir, "00-files.csv"))
			if err != nil {
				return err
			}
			s.filedWriter = f
		}
		s.filedCsvWriter = csv.NewWriter(s.filedWriter)
		err := s.filedCsvWriter.Write([]string{"path", "size", "modified"})
		if err != nil {
			return err
		}
	}
	fileInfo, err := d.Info()
	if err != nil {
		return err
	}
	s.metadata.FileCount++
	s.metadata.TotalSize += fileInfo.Size()
	return s.filedCsvWriter.Write([]string{
		path,
		fmt.Sprintf("%d", fileInfo.Size()),
		fileInfo.ModTime().Format(time.RFC3339),
	})
}

func (s *projectScanner) generate() error {
	err := fs.WalkDir(s.fs, ".", func(path1 string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".devbox" || d.Name() == "node_modules" || d.Name() == ".gradle" {
				return fs.SkipDir
			}
			ignoreFile := path.Join(path1, s.ignoreFile)
			ignoreFileStat, err := fs.Stat(s.fs, ignoreFile)
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
		} else {
			if strings.HasSuffix(path1, ".txt") ||
				strings.HasSuffix(path1, ".md") ||
				strings.HasSuffix(path1, ".adoc") {
				targetPath := filepath.Join(s.outputDir, path1)
				err = s.copyFile(path1, targetPath)
				if err != nil {
					return err
				}
			}
			return s.recordFiled(path1, d)
		}
		return nil
	})
	if err != nil {
		return err
	}

	metadataPath := filepath.Join(s.outputDir, "00-metadata.yml")
	metadataFile, err := os.Create(metadataPath)
	if err != nil {
		return err
	}
	defer func(metadataFile *os.File) {
		_ = metadataFile.Close()
	}(metadataFile)

	yamlEncoder := yaml.NewEncoder(metadataFile)
	err = yamlEncoder.Encode(s.metadata)
	if err != nil {
		return err
	}

	return nil
}

func (s *projectScanner) copyFile(src string, dst string) error {
	r, err := s.fs.Open(src)
	if err != nil {
		return err
	}
	defer func(r fs.File) {
		_ = r.Close()
	}(r)

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(w *os.File) {
		_ = w.Close()
	}(w)

	_, err = io.Copy(w, r)
	return err
}
