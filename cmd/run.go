package cmd

import (
	"bytes"
	"context"
	"fmt"

	"github.com/mattn/go-zglob"

	"cuelang.org/go/encoding/gocode/gocodec"

	"github.com/urfave/cli/v2"
)

// Run requests grpc to server using input in test case file.
func Run(c *cli.Context) error {
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
	headers := extractHeaders(c.StringSlice("header"))
	testFiles, err := zglob.Glob(c.Args().Get(1))
	if err != nil {
		return err
	}

	targetTestName := ""
	if c.NArg() > 2 {
		targetTestName = c.Args().Get(2)
	}

	for _, testFile := range testFiles {
		if err := run(serverName, testFile, targetTestName, headers); err != nil {
			return err
		}
	}

	return nil
}

func run(serverHost, testFile, testName string, headers map[string]string) error {
	ins, err := readCueInstance(testFile)
	if err != nil {
		return err
	}

	insVal, _ := ins.Value().Struct()
	cases, _ := insVal.FieldByName("cases", false)

	codec := gocodec.New(&r, &gocodec.Config{})

	testCases := []testCase{}
	err = codec.Encode(cases.Value, &testCases)
	if err != nil {
		return err
	}

	fmt.Println(testFile)
	for _, c := range testCases {
		if testName != "" && testName != c.Name {
			continue
		}
		fmt.Printf("\ttest name: %s\n", c.Name)
		fmt.Printf("\tmethod: %s\n", c.Method)

		h := mergeMap(headers, c.Headers)

		res := &bytes.Buffer{}
		err = invokeRPC(context.Background(), serverHost, c.Method, h, c.Proto, c.ImportPath, c.Input, res)
		if err != nil {
			return fmt.Errorf("invoke grpc: %w", err)
		}
		fmt.Println("\toutput:", addTabToNewline(res.String(), 2))
	}

	return nil
}
