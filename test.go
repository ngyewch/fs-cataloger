package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/urfave/cli/v2"
	"os"
)

func doTest(cCtx *cli.Context) error {
	token := githubTokenFlag.Get(cCtx)

	var auth transport.AuthMethod
	if token != "" {
		auth = &http.BasicAuth{
			Username: "github", // NOTE this can be anything except an empty string
			Password: token,
		}
	}
	r, err := git.PlainClone("output2", false, &git.CloneOptions{
		Auth:     auth,
		URL:      "https://github.com/ngyewch/data-catalog-test.git",
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	ref, err := r.Head()
	if err != nil {
		return err
	}
	fmt.Println(ref.Hash())

	return nil
}
