package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"encoding/json"

	"github.com/vetcher/godecl"
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
	file, err := godecl.ParseFile(f)
	if err != nil {
		fmt.Println(err)
	}
	t, err := json.Marshal(file)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(t))
}
