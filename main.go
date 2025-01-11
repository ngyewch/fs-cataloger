package main

import (
	"github.com/ngyewch/fs-cataloger/cataloger"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"time"
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

	timeLocation, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		return err
	}

	options := cataloger.Options{
		AnchorFile:      anchorFile,
		IgnoreFile:      ignoreFile,
		OutputDir:       outputDir,
		BaseDirectories: baseDirectories,
		TimeLocation:    timeLocation,
	}
	err = cataloger.Generate(options)
	if err != nil {
		return err
	}

	return nil
}
