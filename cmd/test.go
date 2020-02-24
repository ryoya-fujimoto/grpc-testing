package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-cmp/cmp"

	"github.com/mattn/go-zglob"

	"github.com/kylelemons/godebug/pretty"

	"cuelang.org/go/encoding/gocode/gocodec"
	"github.com/urfave/cli/v2"
)

// Test runs test from test case file.
func Test(c *cli.Context) error {
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

	errs := []string{}
	for _, testFile := range testFiles {
		testFails, err := test(serverName, testFile, targetTestName, headers)
		if err != nil {
			return err
		}
		errs = append(errs, testFails...)
	}

	if len(errs) > 0 {
		return fmt.Errorf("test failed")
	}
	return nil
}

func test(serverHost, testFile, testName string, headers map[string]string) ([]string, error) {
	ins, err := readCueInstance(testFile)
	if err != nil {
		return nil, err
	}

	insVal, _ := ins.Value().Struct()
	cases, _ := insVal.FieldByName("cases")

	codec := gocodec.New(&r, &gocodec.Config{})

	testCases := []testCase{}
	err = codec.Encode(cases.Value, &testCases)
	if err != nil {
		return nil, err
	}

	fmt.Println(testFile)
	errs := []string{}
	for _, c := range testCases {
		if testName != "" && testName != c.Name {
			continue
		}
		tName := c.Name
		if tName == "" {
			tName = c.Method
		}

		h := mergeMap(headers, c.Headers)

		res := &bytes.Buffer{}
		invokeRPC(context.Background(), serverHost, c.Method, h, c.Proto, c.ImportPath, c.Input, res)

		expectJSON := map[string]interface{}{}
		resJSON := map[string]interface{}{}

		if err := json.Unmarshal(c.Output, &expectJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(res.Bytes(), &resJSON); err != nil {
			return nil, err
		}

		if diff := compareResult(expectJSON, resJSON); diff != "" {
			ej, _ := json.Marshal(expectJSON)
			rj, _ := json.Marshal(resJSON)
			errs = append(errs, fmt.Sprintf("expect: %s, but: %s", string(ej), string(rj)))
			fmt.Printf("\tNG: %s\n\t\t%s\n", tName, addTabToNewline(pretty.Compare(expectJSON, resJSON), 2))
		} else {
			fmt.Printf("\tOK: %s\n", tName)
		}
	}

	return errs, nil
}

func compareResult(expect, result map[string]interface{}) string {
	mapComparer := cmp.Comparer(func(x, y map[string]interface{}) bool {
		for key := range x {
			if _, ok := y[key]; !ok {
				delete(x, key)
			}
		}
		for key := range y {
			if _, ok := x[key]; !ok {
				delete(y, key)
			}
		}

		return cmp.Equal(x, y)
	})
	filter := cmp.FilterValues(func(x, y map[string]interface{}) bool {
		return len(x) != len(y)
	}, mapComparer)

	return cmp.Diff(expect, result, filter)
}
