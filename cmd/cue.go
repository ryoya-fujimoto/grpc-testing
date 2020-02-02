package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/protobuf"
	"github.com/mattn/go-zglob"
)

var r cue.Runtime

var wellKnowns = map[string]string{
	"google/protobuf/timestamp.proto": "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/timestamp.proto",
}
var wellKnownRoot = "./tmp/wellknowns"

type testCase struct {
	Name       string
	Method     string
	Proto      []string
	ImportPath []string `json:"import_path"`
	Headers    map[string]string
	Input      json.RawMessage
	Output     json.RawMessage
}

func loadSchemasFromProto(protoRoot string, globs []string) (*cue.Instance, error) {
	protoFiles := []string{}
	for _, glob := range globs {
		pFiles, err := zglob.Glob(glob)
		if err != nil {
			return nil, err
		}
		protoFiles = append(protoFiles, pFiles...)
	}

	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no protofiles")
	}

	err := downloadWellKnowns()
	if err != nil {
		return nil, err
	}

	instances := []*cue.Instance{}
	for _, protoFile := range protoFiles {
		p, _ := filepath.Abs(protoFile)
		file, err := protobuf.Extract(p, nil, &protobuf.Config{
			Paths: []string{protoRoot, wellKnownRoot},
		})
		if err != nil {
			return nil, err
		}

		result, err := r.CompileFile(file)
		if err != nil {
			return nil, err
		}
		instances = append(instances, result)
	}

	return cue.Merge(instances...), nil
}

func readCueInstance(filename string) (*cue.Instance, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return r.Compile(filename, file)
}

func downloadWellKnowns() error {
	targets := map[string]string{}
	for key, url := range wellKnowns {
		_, err := os.Stat(filepath.Join(wellKnownRoot, key))
		if err != nil {
			targets[key] = url
		}
	}

	if len(targets) == 0 {
		return nil
	}
	fmt.Println("download well-known types")
	for key, url := range targets {
		p := filepath.Join(wellKnownRoot, key)
		fmt.Printf("download %s from %s\n", key, url)

		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return err
		}

		if err := downloadFile(url, p); err != nil {
			return err
		}
	}

	return nil
}

func downloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
