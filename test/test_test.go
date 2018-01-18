package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/vetcher/godecl"
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
