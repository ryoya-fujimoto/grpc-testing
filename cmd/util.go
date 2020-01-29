package cmd

import (
	"path/filepath"
)

func extractTarget(testName, testDir string) (baseName string, filePath string) {
	baseName = filepath.Base(testName)
	return baseName, filepath.Join(testDir, filepath.Dir(testName), baseName+".cue")
}
