package main

import (
	"fmt"
	"os"

	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/cli"
)

func main() {
	appSvc := app.NewDefault()
	if err := cli.NewRootCmd(appSvc).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCode(err))
	}
}
