package cmd

import (
	"bytes"
	"context"
	"fmt"

	"cuelang.org/go/encoding/gocode/gocodec"

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
	_, testFile := extractTarget(c.Args().Get(1), testDir)

	targetTestName := ""
	if c.NArg() > 2 {
		targetTestName = c.Args().Get(2)
	}

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

	fmt.Println(testFile)
	for _, c := range testCases {
		if targetTestName != "" && targetTestName != c.Name {
			continue
		}
		fmt.Printf("\ttest name: %s\n", c.Name)
		fmt.Printf("\tmethod: %s\n", c.Method)
		res := &bytes.Buffer{}
		invokeRPC(context.Background(), serverName, c.Method, c.Input, res)
		fmt.Println("\toutput:", addTabToNewline(res.String(), 2))
	}

	return nil
}
