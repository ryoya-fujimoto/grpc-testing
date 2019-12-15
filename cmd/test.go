package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"cuelang.org/go/encoding/gocode/gocodec"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"
)

// Test runs test from test case file.
func Test(c *cli.Context) error {
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
	testName := c.Args().Get(1)
	testFile := filepath.Join(testDir, strcase.ToLowerCamel(testName)+".cue")

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

	errs := []string{}
	for _, c := range testCases {
		res := &bytes.Buffer{}
		invokeRPC(context.Background(), serverName, c.Method, c.Input, res)

		expectJSON := map[string]interface{}{}
		resJSON := map[string]interface{}{}

		if err := json.Unmarshal(c.Output, &expectJSON); err != nil {
			return err
		}
		if err := json.Unmarshal(res.Bytes(), &resJSON); err != nil {
			return err
		}

		if !reflect.DeepEqual(expectJSON, resJSON) {
			ej, _ := json.Marshal(expectJSON)
			rj, _ := json.Marshal(resJSON)
			errs = append(errs, fmt.Sprintf("expect: %s, but: %s", string(ej), string(rj)))
		}
	}
	if len(errs) == 0 {
		fmt.Println("OK:", testName)
	} else {
		cli.Exit("NG", 1)
		fmt.Println("NG:", testName)
		for _, err := range errs {
			fmt.Println(err)
		}
	}

	return nil
}
