package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
	"github.com/urfave/cli/v2"
)

// Add test case file
func Add(c *cli.Context) error {
	testDir := "tests"

	if c.NArg() == 0 {
		fmt.Println("Please specify test case name")
		cli.ShowCommandHelpAndExit(c, "add", 1)
		return nil
	}

	targetName, outPath := extractTarget(c.Args().Get(0), testDir)

	protoRoot := c.String("proto_path")
	if protoRoot == "" {
		protoRoot = "./"
	}
	protoFiles := c.StringSlice("protofiles")

	_, err := os.Stat(outPath)
	if err == nil {
		fmt.Printf("%s is already exists", outPath)
		return nil
	}

	err = os.MkdirAll(filepath.Dir(outPath), 0744)
	if err != nil {
		return err
	}

	tpl := template.New("schema")
	tpl.Parse(testCaseSchema)
	m := map[string]string{
		"Name": targetName,
	}
	var base bytes.Buffer
	_ = tpl.Execute(&base, m)

	ins, err := r.Compile(targetName+".cue", base.String())
	if err != nil {
		return err
	}

	schemas, err := loadSchemasFromProto(protoRoot, protoFiles)
	if err != nil {
		if err.Error() == "no protofiles" {
			fmt.Println("No protofiles. Will not generate schemas.")
		} else {
			return err
		}
	} else {
		ins = cue.Merge(schemas, ins)
	}

	err = ins.Value().Validate()
	if err != nil {
		return err
	}
	op := cue.Raw()
	b, _ := format.Node(ins.Value().Syntax(op))

	err = ioutil.WriteFile(outPath, b, 0644)
	if err != nil {
		return err
	}
	fmt.Println("create:", outPath)

	return nil
}

const testCaseSchema = `name: "{{.Name}}"
Input: {}
Output: {}
Test :: {
	name: string
	method: string
	input: Input
	output: Output
}
cases: [...Test] & [
	{
		name: ""
		method: ""
		input: {}
		output: {}
	},
]
`
