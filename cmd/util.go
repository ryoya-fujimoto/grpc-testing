package cmd

import (
	"path/filepath"
	"strings"
)

func extractTarget(testName, testDir string) (baseName string, filePath string) {
	baseName = filepath.Base(testName)
	filename := baseName
	if !strings.HasSuffix(filename, ".cue") {
		filename = filename + ".cue"
	}
	return baseName, filepath.Join(testDir, filepath.Dir(testName), filename)
}

func addTabToNewline(str string, tabNum int) string {
	return strings.Replace(str, "\n", "\n\t\t", -1)
}
