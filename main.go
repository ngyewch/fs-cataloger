package main

import (
	"github.com/ngyewch/fs-cataloger/cataloger"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
)

var (
	anchorFileFlag = &cli.StringFlag{
		Name:  "anchor-file",
		Usage: "anchor file",
		Value: "README.md",
	}
	ignoreFileFlag = &cli.StringFlag{
		Name:  "ignore-file",
		Usage: "ignore file",
		Value: ".catalogignore",
	}
	outputDirFlag = &cli.PathFlag{
		Name:     "output-dir",
		Usage:    "output directory",
		Required: true,
	}

	app = &cli.App{
		Name:      "fs-cataloger",
		Usage:     "FS cataloger",
		ArgsUsage: "(base directory...)",
		Action:    doMain,
		Flags: []cli.Flag{
			anchorFileFlag,
			ignoreFileFlag,
			outputDirFlag,
		},
	}
)

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func doMain(cCtx *cli.Context) error {
	anchorFile := anchorFileFlag.Get(cCtx)
	ignoreFile := ignoreFileFlag.Get(cCtx)
	outputDir := outputDirFlag.Get(cCtx)
	baseDirectories := cCtx.Args().Slice()

	rootDir := "/"
	rootFs := os.DirFS(rootDir)

	for i, baseDirectory := range baseDirectories {
		relativePath, err := filepath.Rel("/", baseDirectory)
		if err != nil {
			return err
		}
		baseDirectories[i] = relativePath
	}

	options := cataloger.Options{
		RootDir:         rootDir,
		FS:              rootFs,
		AnchorFile:      anchorFile,
		IgnoreFile:      ignoreFile,
		OutputDir:       outputDir,
		BaseDirectories: baseDirectories,
	}
	err := cataloger.Generate(options)
	if err != nil {
		return err
	}

	return nil
}
