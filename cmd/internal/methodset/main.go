// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Inspect build-selected export data for public API checks.
// SPDX-AI-Date: 2026-11-24

// Command methodset reads one target's go list export inventory on stdin.
package main

import (
	"encoding/json"
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
	"strings"
)

func methods(t types.Type) []string {
	set := types.NewMethodSet(t)
	result := []string{}
	for i := 0; i < set.Len(); i++ {
		selection := set.At(i)
		if selection.Obj().Exported() {
			result = append(result, selection.Obj().Name()+strings.TrimPrefix(types.TypeString(selection.Type(), nil), "func"))
		}
	}
	return result
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: methodset <import path> <type>")
	}
	archives := map[string]string{}
	decoder := json.NewDecoder(os.Stdin)
	for {
		var pkg struct{ ImportPath, Export string }
		err := decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		archives[pkg.ImportPath] = pkg.Export
	}
	lookup := func(path string) (io.ReadCloser, error) {
		archive := archives[path]
		if archive == "" {
			return nil, fmt.Errorf("no export data for %s", path)
		}
		return os.Open(archive)
	}
	pkg, err := importer.ForCompiler(token.NewFileSet(), "gc", lookup).Import(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	obj, ok := pkg.Scope().Lookup(os.Args[2]).(*types.TypeName)
	if !ok {
		log.Fatal("missing type ", os.Args[2])
	}
	t := obj.Type()
	if err := json.NewEncoder(os.Stdout).Encode(map[string][]string{"value": methods(t), "pointer": methods(types.NewPointer(t))}); err != nil {
		log.Fatal(err)
	}
}
