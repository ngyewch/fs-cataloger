package main

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
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
	sourceTypeFlag = &cli.StringFlag{
		Name:  "source-type",
		Usage: "source type",
		Value: "local",
	}

	synologyBaseUrlFlag = &cli.StringFlag{
		Name:     "synology-base-url",
		Usage:    "Synology base URL",
		Category: "Synology",
		EnvVars:  []string{"SYNOLOGY_BASE_URL"},
	}
	synologyUsernameFlag = &cli.StringFlag{
		Name:     "synology-username",
		Usage:    "Synology username",
		Category: "Synology",
		EnvVars:  []string{"SYNOLOGY_USERNAME"},
	}
	synologyPasswordFlag = &cli.StringFlag{
		Name:     "synology-password",
		Usage:    "Synology password",
		Category: "Synology",
		EnvVars:  []string{"SYNOLOGY_PASSWORD"},
	}

	githubTokenFlag = &cli.StringFlag{
		Name:     "github-token",
		Usage:    "GitHub token",
		Category: "GitHub",
		EnvVars:  []string{"GITHUB_TOKEN"},
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
			sourceTypeFlag,
			synologyBaseUrlFlag,
			synologyUsernameFlag,
			synologyPasswordFlag,
		},
		Commands: []*cli.Command{
			{
				Name:   "test",
				Usage:  "test",
				Action: doTest,
				Flags: []cli.Flag{
					githubTokenFlag,
				},
			},
		},
	}
)

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Print(err)
		var errWithStack *errors.Error
		ok := errors.As(err, &errWithStack)
		if ok {
			fmt.Println(errWithStack.ErrorStack())
		}
	}
}
