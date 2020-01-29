package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"

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
	_, testFile := extractTarget(c.Args().Get(1))

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
	errs := []string{}
	for _, c := range testCases {
		if targetTestName != "" && targetTestName != c.Name {
			continue
		}
		tName := c.Name
		if tName == "" {
			tName = c.Method
		}
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
			fmt.Printf("\tNG: %s\n\t\t%s\n", tName, addTabToNewline(pretty.Compare(expectJSON, resJSON), 2))
		} else {
			fmt.Printf("\tOK: %s\n", tName)
		}
	}

	return fmt.Errorf("test failed")
}
