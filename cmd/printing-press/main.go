package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/mvanhorn/cli-printing-press/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		var exitErr *cli.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(cli.ExitUnknownError)
	}
}
