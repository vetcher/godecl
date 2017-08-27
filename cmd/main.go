package main

import (
	"fmt"
	"go-file-struct"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(currentDir, "./test/service.go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(fmt.Errorf("error when parse file: %v", err))
	}
	ast.Print(fset, f)
	go_file_struct.ParseFile(f)
}
