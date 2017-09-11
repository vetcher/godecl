package types

import "fmt"

type Import struct {
	Docs    []string `json:"docs,omitempty"`
	Package string   `json:"package,omitempty"`
	Alias   string   `json:"alias,omitempty"`
}

func (i Import) String() string {
	return fmt.Sprintf("%s \"%s\"", i.Alias, i.Package)
}

func (i Import) GoString() string {
	return i.String()
}
