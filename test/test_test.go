package test

import (
	"fmt"
	astparser "go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"encoding/json"

	"github.com/vetcher/godecl"
	"github.com/vetcher/godecl/types"
)

func TestParser(t *testing.T) {
	info, err := ParseFile("service.go")
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bytes))
}

func ParseFile(filename string) (*types.File, error) {
	path, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("can not filepath.Abs: %v", err)
	}
	fset := token.NewFileSet()
	tree, err := astparser.ParseFile(fset, path, nil, astparser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error when parse file: %v\n", err)
	}
	info, err := godecl.ParseFile(tree)
	if err != nil {
		return nil, fmt.Errorf("error when parsing info from file: %v\n", err)
	}
	return info, nil
}
