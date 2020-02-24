package main

import (
	"fmt"
	"os"

	"github.com/ryoya-fujimoto/grpc-testing/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "grpc-testing"
	app.Usage = "The scenario based grpc runner and tester tool"

	app.Commands = []*cli.Command{
		{
			Name:      "add",
			Usage:     "add test case file",
			Action:    cmd.Add,
			ArgsUsage: "`test file`",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "proto_path",
					Usage: "Protofiles root path",
				},
				&cli.StringSliceFlag{
					Name:  "protofiles",
					Usage: "Protofiles to include. You can use grob patterns.",
				},
			},
		},
		{
			Name:      "validate",
			Usage:     "validate test case files with schemas",
			Action:    cmd.Validate,
			ArgsUsage: "`test file(can use glob)`",
		},
		{
			Name:      "run",
			Usage:     "Requests grpc to server using input in test case file, and output response.",
			Action:    cmd.Run,
			ArgsUsage: "`hostname` `test file(can use glob)` `[(optional) test name]`",
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:  "header",
					Usage: "Additional RPC headers in 'name: value' format.",
				},
			},
		},
		{
			Name:      "test",
			Usage:     "Requests grpc to server using input in test case file, and compare between response and output parameter.",
			Action:    cmd.Test,
			ArgsUsage: "`hostname` `test file(can use glob)` `[(optional) test name]`",
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:  "header",
					Usage: "Additional RPC headers in 'name: value' format.",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
