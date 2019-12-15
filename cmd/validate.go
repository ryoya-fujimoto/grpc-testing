package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/mattn/go-zglob"
	"github.com/urfave/cli/v2"
)

// Validate validates test files with schemas.
func Validate(c *cli.Context) error {
	testName := "*"
	if c.NArg() > 0 {
		testName = strcase.ToLowerCamel(c.Args().First())
	}
	fName := testName + ".cue"

	testFiles, err := zglob.Glob(filepath.Join("tests", fName))
	if err != nil {
		return err
	}

	for _, testFile := range testFiles {
		ins, err := readCueInstance(testFile)
		if err != nil {
			return err
		}
		if err := ins.Value().Validate(); err != nil {
			fmt.Println("NG:", testFile)
			fmt.Println(err.Error())
		} else {
			fmt.Println("OK:", testFile)
		}
	}

	return nil
}
