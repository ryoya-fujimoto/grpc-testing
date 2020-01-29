package cmd

import (
	"fmt"

	"github.com/mattn/go-zglob"
	"github.com/urfave/cli/v2"
)

// Validate validates test files with schemas.
func Validate(c *cli.Context) error {
	if c.NArg() == 0 {
		fmt.Println("Please specify test file")
		cli.ShowCommandHelpAndExit(c, "validate", 1)
		return nil
	}
	_, testFilePattern := extractTarget(c.Args().Get(1))

	testFiles, err := zglob.Glob(testFilePattern)
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
