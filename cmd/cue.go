package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/parser"

	"cuelang.org/go/cue/format"
	"cuelang.org/go/encoding/protobuf"

	"github.com/emicklei/proto"

	"cuelang.org/go/cue"
	"github.com/mattn/go-zglob"
)

var r cue.Runtime

var wellKnowns = map[string]string{
	"google/protobuf/timestamp.proto":      "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/timestamp.proto",
	"google/protobuf/any.proto":            "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/any.proto",
	"google/protobuf/api.proto":            "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/api.proto",
	"google/protobuf/descriptor.proto":     "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/descriptor.proto",
	"google/protobuf/duration.proto":       "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/duration.proto",
	"google/protobuf/empty.proto":          "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/empty.proto",
	"google/protobuf/field_mask.proto":     "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/field_mask.proto",
	"google/protobuf/source_context.proto": "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/source_context.proto",
	"google/protobuf/struct.proto":         "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/struct.proto",
	"google/protobuf/type.proto":           "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/type.proto",
	"google/protobuf/wrappers.proto":       "https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/wrappers.proto",
}
var wellKnownRoot = "./tmp/wellknowns"

const syntaxVersion = -1000 + 13

var loadConfig = &load.Config{
	Context: build.NewContext(
		build.ParseFile(func(name string, src interface{}) (*ast.File, error) {
			return parser.ParseFile(name, src,
				parser.FromVersion(syntaxVersion),
				parser.ParseComments,
			)
		})),
}

type testCase struct {
	Name       string
	Method     string
	Proto      []string
	ImportPath []string `json:"import_path"`
	Headers    map[string]string
	Input      json.RawMessage
	Output     json.RawMessage
}

func generateCUEModule(protoRoot string, globs []string) ([]string, error) {
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

	cueImports := []string{}
	for _, protoFile := range protoFiles {
		pkg, _, err := extractProto(protoFile)
		if err != nil {
			return nil, err
		}
		cueImports = append(cueImports, strings.ReplaceAll(pkg, ";", ":"))
	}

	moduleName := ""
	if len(cueImports) > 0 {
		mp := strings.Split(cueImports[0], "/")[:2]
		moduleName = filepath.Join(mp...)
		err := generateModuleFiles(moduleName)
		if err != nil {
			return nil, err
		}
	}

	// generate cue files
	for _, protoFile := range protoFiles {
		_, err := generateCUEFile(moduleName, protoRoot, protoFile)
		if err != nil {
			return nil, err
		}
	}

	return cueImports, nil
}

func generateCUEFile(moduleName, protoRoot, protoFile string) (string, error) {
	p := protoFile
	if !strings.HasPrefix(p, protoRoot) {
		p = filepath.Join(protoRoot, protoFile)
	}

	pkg, imports, err := extractProto(p)
	if err != nil {
		return "", err
	}

	pkgDic := map[string]string{}
	for _, imp := range imports {
		_, ok := wellKnowns[imp]
		if ok {
			continue
		}

		pkgName, err := generateCUEFile(moduleName, protoRoot, imp)
		if err != nil {
			return "", err
		}
		pkgDic[strings.Split(pkgName, ";")[0]] = strings.Split(pkgName, ";")[1]
	}
	fmt.Printf("generate cue file from: %s\n", p)

	cueFile, err := protobuf.Extract(p, nil, &protobuf.Config{
		Paths: []string{protoRoot, wellKnownRoot},
	})
	if err != nil {
		return "", err
	}

	for _, imp := range cueFile.Imports {
		impPath := strings.ReplaceAll(imp.Path.Value, "\"", "")
		if pkgName, ok := pkgDic[impPath]; ok {
			imp.Path.Value = fmt.Sprintf("\"%s:%s\"", impPath, pkgName)
		}
	}

	outDir := strings.ReplaceAll(strings.Split(pkg, ";")[0], moduleName+"/", "")
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return pkg, err
	}

	outPath := filepath.Join(outDir, filepath.Base(cueFile.Filename))
	outFile, err := os.Create(outPath)
	if err != nil {
		return pkg, err
	}
	defer outFile.Close()

	b, err := format.Node(cueFile)
	if err != nil {
		return pkg, err
	}

	_, err = outFile.Write(b)
	if err != nil {
		return pkg, err
	}

	return pkg, nil
}

func generateModuleFiles(moduleName string) error {
	err := os.MkdirAll("./cue.mod/pkg", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll("./cue.mod/usr", 0755)
	if err != nil {
		return err
	}

	p, err := os.Create("./cue.mod/module.cue")
	if err != nil {
		return err
	}
	defer p.Close()

	_, err = p.WriteString(fmt.Sprintf("module: \"%s\"", moduleName))
	return err
}

func extractProto(filePath string) (pkgName string, imports []string, err error) {
	r, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer r.Close()

	parser := proto.NewParser(r)
	def, err := parser.Parse()
	if err != nil {
		return "", nil, err
	}

	for _, e := range def.Elements {
		switch x := e.(type) {
		case *proto.Option:
			pkgName = x.Constant.Source
		case *proto.Import:
			imports = append(imports, x.Filename)
		}
	}

	return pkgName, imports, nil
}

func readCueInstance(filename string) (*cue.Instance, error) {
	binsts := load.Instances([]string{filename}, loadConfig)
	if len(binsts) == 0 {
		return nil, fmt.Errorf("not found cue file")
	}

	insts := []*cue.Instance{}
	for _, binst := range binsts {
		ins, err := r.Build(binst)
		if err != nil {
			return nil, err
		}
		insts = append(insts, ins)
	}

	return cue.Merge(insts...), nil
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
