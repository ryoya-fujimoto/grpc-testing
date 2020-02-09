package cmd

import (
	"bytes"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/urfave/cli/v2"
)

// Add test case file
func Add(c *cli.Context) error {
	if c.NArg() == 0 {
		fmt.Println("Please specify test case name")
		cli.ShowCommandHelpAndExit(c, "add", 1)
		return nil
	}

	targetName, outPath := extractTarget(c.Args().Get(0))

	protoRoot := c.String("proto_path")
	if protoRoot == "" {
		protoRoot = "./"
	}
	protoFiles := c.StringSlice("protofiles")

	// _, err := os.Stat(outPath)
	// if err == nil {
	// 	fmt.Printf("%s is already exists", outPath)
	// 	return nil
	// }

	cueImports, mergeInstances, err := generateCUEModule(protoRoot, protoFiles)
	if err != nil {
		if err.Error() == "no protofiles" {
			fmt.Println("No protofiles. Will not generate schemas.")
		} else {
			return err
		}
	}

	tpl := template.New("schema")
	tpl.Parse(testCaseSchema)
	m := map[string]interface{}{
		"Name":    targetName,
		"Imports": cueImports,
	}
	var base bytes.Buffer
	_ = tpl.Execute(&base, m)
	fmt.Println("hogehoge3")
	testInstance, err := r.Compile(outPath, base.Bytes())
	if err != nil {
		return err
	}

	// err = ioutil.WriteFile(outPath, base.Bytes(), 0644)
	// if err != nil {
	// 	return err
	// }

	fmt.Println("hogehoge1")
	// testInstance, err := readCueInstance(outPath)
	// if err != nil {
	// 	return err
	// }

	fmt.Println("hogehoge2")
	if len(mergeInstances) > 0 {
		mergeInstances = append(mergeInstances, testInstance)
		testInstance = cue.Merge(mergeInstances...)
	}

	op := cue.Raw()
	b, err := format.Node(testInstance.Value().Syntax(op))
	if err != nil {
		return err
	}

	fmt.Println("hogehoge3")
	err = ioutil.WriteFile(outPath, b, 0644)
	if err != nil {
		return err
	}

	fmt.Println("create:", outPath)
	return nil
}

const testCaseSchema = `{{range .Imports}}import "{{.}}"
{{end}}
name: "{{.Name}}"
Input: {}
Output: {}
Test :: {
	name: string
	method: string
	proto?: [...string]
	import_path?: [...string]
	headers?: [string]: string
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
