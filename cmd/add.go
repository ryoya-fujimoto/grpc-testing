package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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

	_, err := os.Stat(outPath)
	if err == nil {
		fmt.Printf("%s is already exists", outPath)
		return nil
	}

	cueImports, _, err := generateCUEModule(protoRoot, protoFiles)
	if err != nil {
		if err.Error() == "no protofiles" {
			fmt.Println("No protofiles. Will not generate schemas.")
		} else {
			return err
		}
	} else {
		fmt.Println(cueImports)
	}

	tpl := template.New("schema")
	tpl.Parse(testCaseSchema)
	m := map[string]interface{}{
		"Name":    targetName,
		"Imports": cueImports,
	}
	var base bytes.Buffer
	_ = tpl.Execute(&base, m)

	// TODO: merge non package instance

	// testCUE, err := parser.ParseFile(targetName+".cue", base.String())
	// testInstance, err := r.Parse(targetName+".cue", base.String())
	// if err != nil {
	// 	return err
	// }
	// buildIns := &build.Instance{
	// 	Files: []*ast.File{testCUE},
	// }
	// testInstance, err := r.Build(buildIns)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println("hoge1----")
	// if len(mergeInstances) > 0 {
	// 	mergeInstances = append(mergeInstances, testInstance)
	// 	testInstance = cue.Merge(mergeInstances...)
	// }

	// op := cue.Raw()
	// b, err := format.Node(testInstance.Value().Syntax(op))
	// if err != nil {
	// 	return err
	// }

	err = ioutil.WriteFile(outPath, base.Bytes(), 0644)
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
