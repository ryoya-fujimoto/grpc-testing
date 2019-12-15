package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/protobuf"
	"github.com/mattn/go-zglob"
)

var r cue.Runtime

type testCase struct {
	Method string
	Input  json.RawMessage
	Output json.RawMessage
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

	instances := []*cue.Instance{}
	for _, protoFile := range protoFiles {
		p, _ := filepath.Abs(protoFile)
		file, err := protobuf.Extract(p, nil, &protobuf.Config{
			Paths: []string{protoRoot},
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
