package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"io/ioutil"

	"path/filepath"

	"github.com/vetcher/godecl"
)

const (
	source    = "source.go"
	result    = "result.json"
	assetsDir = "assets"
)

func TestParser(t *testing.T) {
	info, err := godecl.ParseFile("service.go")
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bytes))
}

type AstraTest struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func TestParsingFile(t *testing.T) {
	d, err := ioutil.ReadFile("./list_of_tests.json")
	if err != nil {
		t.Fatal(err)
	}
	var tests []AstraTest
	err = json.Unmarshal(d, &tests)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			expected, err := ioutil.ReadFile(filepath.Join(assetsDir, tt.Path, result))
			if err != nil {
				t.Fatal(err)
			}
			file, err := godecl.ParseFile(filepath.Join(assetsDir, tt.Path, source))
			if err != nil {
				t.Fatal(err)
			}
			actual, err := json.Marshal(file)
			if err != nil {
				t.Fatal(err)
			}
			if !testEq(expected, actual) {
				t.Fatalf("expected != actual:\n%s\n\n%s", string(expected), string(actual))
			}
		})
	}
}

func testEq(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
