package cmd

import (
	"path/filepath"
)

func extractTarget(testName string) (string, string) {
	return filepath.Base(testName), filepath.Dir(testName)
}
