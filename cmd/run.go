package cmd

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"cuelang.org/go/encoding/gocode/gocodec"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"
)

// Run requests grpc to server using input in test case file.
func Run(c *cli.Context) error {
	testDir := "tests"

	if c.NArg() == 0 {
		fmt.Println("Please specify server name")
		cli.ShowCommandHelpAndExit(c, "run", 1)
		return nil
	}
	serverName := c.Args().Get(0)
	if c.NArg() == 1 {
		fmt.Println("Please specify test name")
		cli.ShowCommandHelpAndExit(c, "run", 1)
		return nil
	}
	targetName, targetDir := extractTarget(c.Args().Get(1))
	testFile := filepath.Join(testDir, targetDir, strcase.ToLowerCamel(targetName)+".cue")

	ins, err := readCueInstance(testFile)
	if err != nil {
		return err
	}

	insVal, _ := ins.Value().Struct()
	cases, _ := insVal.FieldByName("cases")

	codec := gocodec.New(&r, &gocodec.Config{})

	testCases := []testCase{}
	err = codec.Encode(cases.Value, &testCases)
	if err != nil {
		return err
	}

	for _, c := range testCases {
		res := &bytes.Buffer{}
		invokeRPC(context.Background(), serverName, c.Method, c.Input, res)
		fmt.Println("output:", res.String())
	}

	return nil
}
