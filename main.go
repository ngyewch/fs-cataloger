package main

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/ngyewch/fs-cataloger/cataloger"
	"github.com/ngyewch/go-syno/api"
	"github.com/ngyewch/go-syno/api/auth"
	synoFs "github.com/ngyewch/go-syno/fs"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
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

func doMain(cCtx *cli.Context) error {
	anchorFile := anchorFileFlag.Get(cCtx)
	ignoreFile := ignoreFileFlag.Get(cCtx)
	outputDir := outputDirFlag.Get(cCtx)
	sourceType := sourceTypeFlag.Get(cCtx)
	baseDirectories := cCtx.Args().Slice()

	var synologyClient *api.Client
	if sourceType == "synology" {
		baseUrl := synologyBaseUrlFlag.Get(cCtx)
		username := synologyUsernameFlag.Get(cCtx)
		password := synologyPasswordFlag.Get(cCtx)

		c, err := api.NewClient(baseUrl, &http.Client{})
		if err != nil {
			return err
		}
		authApi, err := auth.New(c)
		if err != nil {
			return err
		}

		sessionId := uuid.New().String()

		loginResponse, err := authApi.Login(auth.LoginRequest{
			Account: username,
			Passwd:  password,
			Session: sessionId,
		})
		if err != nil {
			return err
		}

		c.SetParam("_sid", loginResponse.Data.Sid)

		defer func() {
			_ = func() error {
				_, err := authApi.Logout(auth.LogoutRequest{
					Session: sessionId,
				})
				return err
			}
		}()

		synologyClient = c
	}

	var sources []cataloger.Source
	for _, baseDirectory := range baseDirectories {
		switch sourceType {
		case "local":
			sources = append(sources, cataloger.Source{
				RootDir: baseDirectory,
				FS:      os.DirFS(baseDirectory),
			})

		case "synology":
			sourceFs, err := synoFs.NewFS(synologyClient, baseDirectory)
			if err != nil {
				return err
			}
			sources = append(sources, cataloger.Source{
				RootDir: baseDirectory,
				FS:      sourceFs,
			})

		default:
			return fmt.Errorf("unsupported source type: %s", sourceType)
		}
	}

	options := cataloger.Options{
		AnchorFile: anchorFile,
		IgnoreFile: ignoreFile,
		OutputDir:  outputDir,
		Sources:    sources,
	}
	err := cataloger.Generate(options)
	if err != nil {
		return err
	}

	return nil
}
