package main

import (
	"fmt"
	"os"

	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/cli"
	"github.com/zhongyangchuwu/shelf-go/internal/jsonvault"
)

func main() {
	appSvc := app.New(jsonvault.Provider{})
	if err := cli.NewRootCmd(appSvc).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCode(err))
	}
}
