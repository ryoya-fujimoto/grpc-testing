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
	cases, _ := insVal.FieldByName("cases", false)

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

		tl := []testData{}
		for _, td := range c.Tests {
			tl = append(tl, testData{
				name:    tName,
				method:  c.Method,
				headers: h,
				input:   td.Input,
				output:  td.Output,
			})
		}

		if len(tl) == 0 {
			tl = append(tl, testData{
				name:    tName,
				method:  c.Method,
				headers: h,
				input:   c.Input,
				output:  c.Output,
			})
		}

		es, err := testList(
			context.Background(),
			serverHost,
			c.Proto,
			c.ImportPath,
			tl,
		)
		if err != nil {
			return nil, err
		}
		errs = append(errs, es...)
	}

	return errs, nil
}

func testList(ctx context.Context, serverHost string, proto, importPath []string, tdList []testData) ([]string, error) {
	errs := []string{}

	for _, td := range tdList {
		res := &bytes.Buffer{}
		invokeRPC(ctx, serverHost, td.method, td.headers, proto, importPath, td.input, res)

		expectJSON := map[string]interface{}{}
		resJSON := map[string]interface{}{}

		if err := json.Unmarshal(td.output, &expectJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(res.Bytes(), &resJSON); err != nil {
			return nil, err
		}

		if diff := compareResult(expectJSON, resJSON); diff != "" {
			ej, _ := json.Marshal(expectJSON)
			rj, _ := json.Marshal(resJSON)
			errs = append(errs, fmt.Sprintf("expect: %s, but: %s", string(ej), string(rj)))
			fmt.Printf("\tNG: %s\n\t\t%s\n", td.name, addTabToNewline(pretty.Compare(expectJSON, resJSON), 2))
		} else {
			fmt.Printf("\tOK: %s\n", td.name)
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
