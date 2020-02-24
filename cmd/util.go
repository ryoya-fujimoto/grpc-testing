package cmd

import (
	"path/filepath"
	"strings"
)

func extractTarget(testName string) (baseName string, filePath string) {
	baseName = filepath.Base(testName)
	filename := baseName
	if !strings.HasSuffix(filename, ".cue") {
		filename = filename + ".cue"
	}
	return baseName, filepath.Join(filepath.Dir(testName), filename)
}

func addTabToNewline(str string, tabNum int) string {
	return strings.Replace(str, "\n", "\n\t\t", -1)
}

func extractHeaders(headers []string) map[string]string {
	h := map[string]string{}
	for _, headerSet := range headers {
		s := strings.Split(headerSet, ":")
		if len(s) < 2 {
			continue
		}
		h[strings.TrimSpace(s[0])] = strings.TrimSpace(s[1])
	}

	return h
}

func mergeMap(m1, m2 map[string]string) map[string]string {
	result := map[string]string{}

	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return (result)
}
