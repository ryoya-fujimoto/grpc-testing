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

		h := mergeMap(headers, c.Headers)

		testList := []testData{}
		for _, td := range c.Tests {
			testList = append(testList, testData{
				name:    c.Name,
				method:  c.Method,
				headers: h,
				input:   td.Input,
				output:  td.Output,
			})
		}

		if len(testList) == 0 {
			testList = append(testList, testData{
				name:    c.Name,
				method:  c.Method,
				headers: h,
				input:   c.Input,
				output:  c.Output,
			})
		}

		err := runList(
			context.Background(),
			serverHost,
			c.Proto,
			c.ImportPath,
			testList,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func runList(ctx context.Context, serverHost string, proto, importPath []string, tdList []testData) error {
	for _, td := range tdList {
		fmt.Printf("\ttest name: %s\n", td.name)
		fmt.Printf("\tmethod: %s\n", td.method)

		res := &bytes.Buffer{}

		err := invokeRPC(
			context.Background(),
			serverHost,
			td.method,
			td.headers,
			proto,
			importPath,
			td.input,
			res)
		if err != nil {
			return fmt.Errorf("invoke grpc: %w", err)
		}
		fmt.Println("\toutput:", addTabToNewline(res.String(), 2))
	}

	return nil
}
